"""CLI entry point for pm."""

import os
import subprocess
from datetime import date
from pathlib import Path

import click

from pm.config import PROJECT_FILES, VALID_STATUSES, load_config, save_config
from pm.project import find_project, scan_all, write_frontmatter


ALIASES = {
    "sw": "switch",
}


class AliasGroup(click.Group):
    def get_command(self, ctx, cmd_name):
        cmd_name = ALIASES.get(cmd_name, cmd_name)
        return super().get_command(ctx, cmd_name)

    def resolve_command(self, ctx, args):
        if args and args[0] in ALIASES:
            args[0] = ALIASES[args[0]]
        return super().resolve_command(ctx, args)


@click.group(cls=AliasGroup)
def cli():
    """pm — project metadata manager."""


# --- Collection commands ---

@cli.group()
def collection():
    """Manage project collections."""


@collection.command("add")
@click.argument("name")
@click.argument("path", type=click.Path())
def collection_add(name: str, path: str):
    """Register a collection root directory."""
    config = load_config()
    resolved = str(Path(path).resolve())
    config.setdefault("collections", {})[name] = resolved
    save_config(config)
    click.echo(f"Collection '{name}' -> {resolved}")


@collection.command("remove")
@click.argument("name")
def collection_remove(name: str):
    """Remove a collection (does not delete files)."""
    config = load_config()
    if name not in config.get("collections", {}):
        click.echo(f"Error: collection '{name}' not found.", err=True)
        raise SystemExit(1)
    del config["collections"][name]
    save_config(config)
    click.echo(f"Collection '{name}' removed.")


@collection.command("list")
def collection_list():
    """List registered collections."""
    config = load_config()
    collections = config.get("collections", {})
    if not collections:
        click.echo("No collections registered.")
        return
    for name, path in collections.items():
        click.echo(f"{name}: {path}")


# --- Project commands ---

@cli.command()
@click.option("--name", "-n", required=True, help="Project name.")
@click.option("--collection", "-c", required=True, help="Collection to create project in.")
@click.option("--tags", "-t", default="", help="Space-separated tags.")
@click.option("--status", "-s", default="active", type=click.Choice(VALID_STATUSES), help="Initial status.")
def init(name: str, collection: str, tags: str, status: str):
    """Initialize a new project under a collection."""
    config = load_config()
    collections = config.get("collections", {})
    if collection not in collections:
        click.echo(f"Error: collection '{collection}' not found. Register it first with: pm collection add", err=True)
        raise SystemExit(1)

    project_dir = Path(collections[collection]) / name
    if project_dir.exists():
        click.echo(f"Error: {project_dir} already exists.", err=True)
        raise SystemExit(1)

    project_dir.mkdir(parents=True)

    meta = {
        "name": name,
        "status": status,
        "start-date": date.today().isoformat(),
        "tags": tags,
    }
    write_frontmatter(project_dir / "README.md", meta)

    for fname in PROJECT_FILES:
        (project_dir / fname).touch()

    click.echo(f"Project '{name}' initialized at {project_dir}")


@cli.command()
@click.argument("name")
@click.option("--path", "-p", default=None, type=click.Path(exists=True), help="Working directory. Defaults to cwd.")
def link(name: str, path: str | None):
    """Link a working directory to a project."""
    project = find_project(name)
    if not project:
        click.echo(f"Error: project '{name}' not found.", err=True)
        raise SystemExit(1)

    work_dir = Path(path).resolve() if path else Path.cwd()

    for fname in PROJECT_FILES:
        src = project.path / fname
        dst = work_dir / fname
        if dst.exists() or dst.is_symlink():
            dst.unlink()
        dst.symlink_to(src)

    # Update workspace in frontmatter
    readme = project.path / "README.md"
    from pm.project import parse_frontmatter
    meta = parse_frontmatter(readme)
    meta["workspace"] = str(work_dir)
    write_frontmatter(readme, meta)

    click.echo(f"Linked '{name}' workspace to {work_dir}")


@cli.command()
@click.argument("name")
@click.argument("file", type=click.Choice(["notes", "journal"]), default="notes")
def edit(name: str, file: str):
    """Edit a project file (notes or journal)."""
    project = find_project(name)
    if not project:
        click.echo(f"Error: project '{name}' not found.", err=True)
        raise SystemExit(1)

    target = project.path / f"{file}.pm.md"
    if not target.exists():
        click.echo(f"Error: {target} does not exist.", err=True)
        raise SystemExit(1)

    editor = os.environ.get("EDITOR", "vim")
    subprocess.run([editor, str(target)])


@cli.command()
@click.argument("name")
@click.argument("target", type=click.Choice(["meta", "work"]), default="work")
def jump(name: str, target: str):
    """Print the path to a project directory (for use with shell wrapper).

    Defaults to the workspace. Use 'meta' to jump to the meta directory.
    """
    project = find_project(name)
    if not project:
        click.echo(f"Error: project '{name}' not found.", err=True)
        raise SystemExit(1)

    if target == "meta":
        d = str(project.path)
    else:
        if not project.workspace:
            click.echo(f"Error: project '{name}' has no linked workspace.", err=True)
            raise SystemExit(1)
        d = project.workspace

    if not Path(d).is_dir():
        click.echo(f"Error: {d} does not exist.", err=True)
        raise SystemExit(1)

    click.echo(d)


@cli.command("switch")
def switch_dir():
    """Switch between meta and workspace for the current project.

    Detects the current project from cwd and prints the other directory.
    """
    cwd = str(Path.cwd())
    for project in scan_all():
        meta = str(project.path)
        work = project.workspace

        if cwd == meta or cwd.startswith(meta + "/"):
            if not work:
                click.echo(f"Error: project '{project.name}' has no linked workspace.", err=True)
                raise SystemExit(1)
            if not Path(work).is_dir():
                click.echo(f"Error: {work} does not exist.", err=True)
                raise SystemExit(1)
            click.echo(work)
            return

        if work and (cwd == work or cwd.startswith(work + "/")):
            click.echo(meta)
            return

    click.echo("Error: current directory is not inside any known project.", err=True)
    raise SystemExit(1)


@cli.command("list")
@click.option("--all", "show_all", is_flag=True, help="Include done projects.")
@click.option("--tags", "-t", default=None, help="Filter by tags (space-separated, OR logic).")
@click.option("--collection", "-c", default=None, help="Filter by collection.")
def list_projects(show_all: bool, tags: str | None, collection: str | None):
    """List projects."""
    projects = scan_all()

    if not show_all:
        projects = [p for p in projects if p.status != "done"]

    if collection:
        projects = [p for p in projects if p.collection == collection]

    if tags:
        filter_tags = set(tags.split())
        projects = [p for p in projects if filter_tags & set(p.tags)]

    if not projects:
        click.echo("No projects found.")
        return

    for p in projects:
        tag_str = f" [{' '.join(p.tags)}]" if p.tags else ""
        work_str = f"  work: {p.workspace}" if p.workspace else ""
        click.echo(f"{p.name} ({p.collection}) [{p.status}]{tag_str}")
        click.echo(f"  meta: {p.path}")
        if work_str:
            click.echo(work_str)


@cli.command("setup_shell")
def setup_shell():
    """Print shell integration code. Usage: eval "$(pm setup_shell)" """
    shell_script = Path(__file__).parent / "shell.zsh"
    click.echo(shell_script.read_text())
