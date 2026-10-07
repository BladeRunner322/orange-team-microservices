#!/bin/bash
#
# PostgreSQL backup script for Orange Team Microservices.
#
# Dumps all service databases (auth, profiles, exercises) into a
# timestamped directory under BACKUP_DIR, then rotates old backups.
#
# Designed to run from cron on the production server. The script is
# kept in git (see scripts/), and the server runs it from the cloned
# repository directory, so `git pull` on deploy keeps it up to date.
#
# Usage:
#   ./scripts/backup-db.sh
#
# Logs to stdout/stderr — redirect in cron.

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/root/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
DATE=$(date +%Y-%m-%d_%H-%M-%S)
TARGET_DIR="${BACKUP_DIR}/${DATE}"

# Off-site (ADR-014). Пустое значение RCLONE_REMOTE — off-site отключается.
RCLONE_REMOTE="${RCLONE_REMOTE:-selectel}"
S3_BUCKET="${S3_BUCKET:-orange-team-backups}"
S3_RETENTION_DAYS="${S3_RETENTION_DAYS:-30}"

# Список сервисов, у которых есть PostgreSQL.
# Контейнер называется <service>-postgres, БД — <service>_db.
# При добавлении нового сервиса (Habits, Workouts, Leaderboard) —
# просто дописать имя в массив.
SERVICES=(auth profiles exercises)

mkdir -p "${TARGET_DIR}"

echo "[$(date)] Backup started → ${TARGET_DIR}"

for svc in "${SERVICES[@]}"; do
  container="${svc}-postgres"
  db="${svc}_db"
  echo "[$(date)] Dumping ${db} from ${container}"

  if ! docker exec "${container}" pg_dump -U test -d "${db}" | gzip > "${TARGET_DIR}/${db}.sql.gz"; then
    echo "[$(date)] ERROR: failed to dump ${db}" >&2
    exit 1
  fi

  # Fail loudly if the dump is suspiciously small (e.g. < 100 bytes gzipped)
  size=$(stat -c%s "${TARGET_DIR}/${db}.sql.gz")
  if [ "${size}" -lt 100 ]; then
    echo "[$(date)] ERROR: dump of ${db} is suspiciously small (${size} bytes)" >&2
    exit 1
  fi
done

echo "[$(date)] Dumps complete"

# Off-site upload (ADR-014). Ошибка выгрузки НЕ фейлит локальный бэкап —
# локальный успех важнее. Пишем WARN и продолжаем.
if [ -n "${RCLONE_REMOTE}" ] && command -v rclone >/dev/null 2>&1; then
  echo "[$(date)] Uploading to s3://${S3_BUCKET} (remote: ${RCLONE_REMOTE})"

  for svc in "${SERVICES[@]}"; do
    db="${svc}_db"
    remote_path="${RCLONE_REMOTE}:${S3_BUCKET}/${svc}/${DATE}/"

    if rclone copy "${TARGET_DIR}/${db}.sql.gz" "${remote_path}" \
         --s3-no-check-bucket --quiet; then
      echo "[$(date)] Uploaded ${db}.sql.gz → ${remote_path}"
    else
      echo "[$(date)] WARN: failed to upload ${db}.sql.gz to S3" >&2
    fi
  done

  # Ротация в S3: удаляем папки старше S3_RETENTION_DAYS.
  for svc in "${SERVICES[@]}"; do
    rclone delete --min-age "${S3_RETENTION_DAYS}d" \
      "${RCLONE_REMOTE}:${S3_BUCKET}/${svc}/" --quiet \
      || echo "[$(date)] WARN: S3 rotation failed for ${svc}" >&2
  done
else
  echo "[$(date)] Off-site disabled (RCLONE_REMOTE empty or rclone not installed)"
fi

# Rotate: remove directories older than RETENTION_DAYS
if [ -d "${BACKUP_DIR}" ]; then
  find "${BACKUP_DIR}" -mindepth 1 -maxdepth 1 -type d -mtime +"${RETENTION_DAYS}" -print -exec rm -rf {} +
fi

echo "[$(date)] Backup finished"