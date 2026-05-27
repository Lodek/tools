"""Configuration management for pm."""

import json
from pathlib import Path

PM_DIR = Path.home() / ".pm"
CONFIG_FILE = PM_DIR / "config.json"

PROJECT_FILES = ["notes.pm.md", "journal.pm.md"]

VALID_STATUSES = ["active", "ongoing", "archived", "done"]


def ensure_dirs():
    PM_DIR.mkdir(parents=True, exist_ok=True)


def load_config() -> dict:
    if CONFIG_FILE.exists():
        return json.loads(CONFIG_FILE.read_text())
    return {"collections": {}}


def save_config(config: dict):
    ensure_dirs()
    CONFIG_FILE.write_text(json.dumps(config, indent=2) + "\n")
