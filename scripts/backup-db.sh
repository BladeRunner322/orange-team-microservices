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

# Rotate: remove directories older than RETENTION_DAYS
if [ -d "${BACKUP_DIR}" ]; then
  find "${BACKUP_DIR}" -mindepth 1 -maxdepth 1 -type d -mtime +"${RETENTION_DAYS}" -print -exec rm -rf {} +
fi

echo "[$(date)] Backup finished"