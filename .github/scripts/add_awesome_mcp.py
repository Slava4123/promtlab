#!/usr/bin/env python3
"""Idempotent insertion of PromptVault entry into punkpeye/awesome-mcp-servers README.

Inserts one line under "### 🧠 Knowledge & Memory" (category header), keeping
alphabetical order by server name. Exits 0 (without changes) if the entry is
already present.

Usage:
    add_awesome_mcp.py <path-to-awesome-mcp-servers-repo>
"""
import re
import sys
from pathlib import Path

ENTRY_NAME = "Slava4123/promptvault"
ENTRY_LINE = (
    "- [Slava4123/promptvault](https://github.com/Slava4123/promptvault) 🏎️ ☁️ "
    "- ПромтЛаб — self-hosted AI prompt library with collections, tags, versioning "
    "and team collaboration. Remote MCP via Streamable HTTP."
)
# Заголовок в punkpeye/awesome-mcp-servers имеет вид:
#   "### 🧠 <a name="knowledge--memory"></a>Knowledge & Memory"
# Inline HTML anchor tag между эмодзи и названием игнорируем. Дефис/длинное
# тире между эмодзи и словами тоже допустим (другие категории их используют).
CATEGORY_HEADER_RE = re.compile(
    r'^###\s*🧠\s*(?:<a[^>]*></a>)?\s*[-–—]?\s*Knowledge\s*&\s*Memory\s*$',
    re.IGNORECASE,
)


def extract_name(line: str) -> str:
    match = re.match(r"^-\s*\[([^\]]+)\]", line)
    return match.group(1).lower() if match else ""


def insert_entry(readme: Path) -> bool:
    lines = readme.read_text(encoding="utf-8").splitlines()

    if any(ENTRY_NAME.lower() in line.lower() for line in lines):
        print(f"entry already present: {ENTRY_NAME}")
        return False

    header_idx = next(
        (i for i, line in enumerate(lines) if CATEGORY_HEADER_RE.match(line.strip())),
        None,
    )
    if header_idx is None:
        raise SystemExit("header '### 🧠 Knowledge & Memory' not found")

    # Scan category body: from header until next h2/h3.
    insert_at = len(lines)
    entry_name_lower = extract_name(ENTRY_LINE)
    for idx in range(header_idx + 1, len(lines)):
        stripped = lines[idx].strip()
        if stripped.startswith("## ") or stripped.startswith("### "):
            insert_at = idx
            break
        if stripped.startswith("- ["):
            if extract_name(stripped) > entry_name_lower:
                insert_at = idx
                break
    else:
        insert_at = len(lines)

    lines.insert(insert_at, ENTRY_LINE)
    readme.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"inserted at line {insert_at + 1}")
    return True


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: add_awesome_mcp.py <repo-path>", file=sys.stderr)
        return 2
    readme = Path(sys.argv[1]) / "README.md"
    if not readme.exists():
        print(f"README.md not found at {readme}", file=sys.stderr)
        return 2
    changed = insert_entry(readme)
    return 0 if changed else 78  # 78 = idempotent no-op sentinel


if __name__ == "__main__":
    sys.exit(main())
