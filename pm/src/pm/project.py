"""Project scanning and frontmatter parsing."""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from pathlib import Path

from pm.config import PROJECT_FILES, VALID_STATUSES, load_config


@dataclass
class Project:
    name: str
    path: Path
    collection: str
    status: str = "active"
    start_date: str = ""
    finish_date: str = ""
    workspace: str = ""
    tags: list[str] = field(default_factory=list)


def parse_frontmatter(readme: Path) -> dict:
    """Parse YAML-like frontmatter from a README.md file."""
    text = readme.read_text()
    match = re.match(r"^---\n(.+?)\n---", text, re.DOTALL)
    if not match:
        return {}

    result = {}
    for line in match.group(1).splitlines():
        if ":" not in line:
            continue
        key, _, value = line.partition(":")
        result[key.strip()] = value.strip()
    return result


def write_frontmatter(readme: Path, meta: dict):
    """Write frontmatter to a README.md, preserving content after the frontmatter."""
    body = ""
    if readme.exists():
        text = readme.read_text()
        match = re.match(r"^---\n.+?\n---\n?(.*)", text, re.DOTALL)
        if match:
            body = match.group(1)
        elif not text.startswith("---"):
            body = text

    lines = ["---"]
    for key, value in meta.items():
        lines.append(f"{key}: {value}")
    lines.append("---")
    if body:
        lines.append(body)
    else:
        lines.append("")

    readme.write_text("\n".join(lines))


def scan_collection(name: str, root: Path) -> list[Project]:
    """Scan a collection directory for projects."""
    projects = []
    if not root.is_dir():
        return projects

    for entry in sorted(root.iterdir()):
        readme = entry / "README.md"
        if not entry.is_dir() or not readme.exists():
            continue

        meta = parse_frontmatter(readme)
        projects.append(Project(
            name=meta.get("name", entry.name),
            path=entry,
            collection=name,
            status=meta.get("status", "active"),
            start_date=meta.get("start-date", ""),
            finish_date=meta.get("finish-date", ""),
            workspace=meta.get("workspace", ""),
            tags=meta.get("tags", "").split() if meta.get("tags") else [],
        ))

    return projects


def scan_all() -> list[Project]:
    """Scan all collections and return all projects."""
    config = load_config()
    projects = []
    for name, root in config.get("collections", {}).items():
        projects.extend(scan_collection(name, Path(root)))
    return projects


def find_project(name: str) -> Project | None:
    """Find a project by name across all collections."""
    for project in scan_all():
        if project.name == name:
            return project
    return None
