import os
import sqlite3
from datetime import datetime, timedelta

DB = os.getenv("FILL_DB", "lilmon.sqlite")
DAYS = os.getenv("FILL_DAYS", "120")
TABLE = os.getenv("FILL_TABLE", "lilmon_metric_n_temp_files")
SHIFT_SEC = os.getenv("FILL_SHIFT_SEC", "15")


if __name__ == "__main__":
    print(f"Operating on {DB}:{TABLE} for {DAYS} days, shift is {SHIFT_SEC} seconds")
    db = sqlite3.connect(DB)
    c = db.cursor()

    t_end = datetime.now()
    t_start = t_end - timedelta(days=int(DAYS))
    print(f"Time range: {t_start} -> {t_end}")

    t = t_start
    q = f"INSERT INTO {TABLE} (value, timestamp) VALUES (?, ?)"
    c.execute("BEGIN")
    i = 0
    while t < t_end:
        c.execute(q, (i, t))
        i += 1
        t += timedelta(seconds=int(SHIFT_SEC))
    c.execute("END")
    db.close()
