from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parent
INDEX_FILE = ROOT / "\u5feb\u7167\u7d22\u5f15.md"

TEXT_SUFFIXES = {
    ".cjs",
    ".css",
    ".html",
    ".js",
    ".json",
    ".md",
    ".mjs",
    ".py",
    ".txt",
}

ALLOWED_EMAILS = {
    "account@example.com",
    "redacted@example.com",
}

FORBIDDEN_PATTERNS = [
    re.compile(r"rongjie\.zhang0714", re.IGNORECASE),
    re.compile("Abc" + "123456", re.IGNORECASE),
    re.compile(r"\u8d26\u53f7\u5bc6\u7801\u5df2\u586b\u5199"),
    re.compile(r"\u8f93\u5165\u8d26\u53f7\u5bc6\u7801"),
]

EMAIL_RE = re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}")
HEX24_RE = re.compile(r"\b[a-f0-9]{24}\b")
MD_LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+)\)")
PNG_REF_RE = re.compile(r"`([^`]+\.png)`")


def rel(path: Path) -> str:
    return path.relative_to(ROOT).as_posix()


def text_files() -> list[Path]:
    return sorted(
        path
        for path in ROOT.rglob("*")
        if path.is_file() and path.suffix.lower() in TEXT_SUFFIXES
    )


def check_utf8(errors: list[str]) -> None:
    for path in text_files():
        try:
            path.read_text(encoding="utf-8")
        except UnicodeDecodeError as exc:
            errors.append(f"utf8: {rel(path)}: {exc}")


def check_markdown_links(errors: list[str]) -> None:
    for path in sorted(ROOT.rglob("*.md")):
        text = path.read_text(encoding="utf-8")
        for match in MD_LINK_RE.finditer(text):
            target = match.group(1).strip("<>")
            if (
                target.startswith("#")
                or target.startswith("http://")
                or target.startswith("https://")
                or target.startswith("mailto:")
            ):
                continue

            target_path = Path(target)
            if not target_path.is_absolute():
                target_path = path.parent / target_path
            if not target_path.exists():
                errors.append(f"link: {rel(path)} -> {target}")


def check_png_index(errors: list[str]) -> None:
    if not INDEX_FILE.exists():
        errors.append(f"index: missing {INDEX_FILE.name}")
        return

    text = INDEX_FILE.read_text(encoding="utf-8")
    refs = PNG_REF_RE.findall(text)
    seen = set(refs)

    for ref in refs:
        if not (ROOT / ref).exists():
            errors.append(f"png-ref: {INDEX_FILE.name} -> {ref}")

    for folder_name in ("snapshots", "litlyx-reference"):
        folder = ROOT / folder_name
        if not folder.exists():
            continue
        for image in sorted(folder.rglob("*.png")):
            image_ref = rel(image)
            if image_ref not in seen:
                errors.append(f"png-unindexed: {image_ref}")


def check_sensitive_text(errors: list[str]) -> None:
    for path in text_files():
        text = path.read_text(encoding="utf-8", errors="ignore")

        for pattern in FORBIDDEN_PATTERNS:
            for match in pattern.finditer(text):
                line = text.count("\n", 0, match.start()) + 1
                errors.append(f"sensitive: {rel(path)}:{line}: {match.group(0)}")

        for match in EMAIL_RE.finditer(text):
            value = match.group(0)
            if value in ALLOWED_EMAILS:
                continue
            line = text.count("\n", 0, match.start()) + 1
            errors.append(f"email: {rel(path)}:{line}: {value}")

        for match in HEX24_RE.finditer(text):
            line = text.count("\n", 0, match.start()) + 1
            errors.append(f"workspace-id-like: {rel(path)}:{line}: {match.group(0)}")


def main() -> int:
    errors: list[str] = []

    check_utf8(errors)
    check_markdown_links(errors)
    check_png_index(errors)
    check_sensitive_text(errors)

    if errors:
        print(f"litlyx asset validation failed: {len(errors)} issue(s)")
        for item in errors:
            print(item)
        return 1

    print("litlyx asset validation passed")
    print(f"text_files={len(text_files())}")
    print(f"indexed_png_refs={len(PNG_REF_RE.findall(INDEX_FILE.read_text(encoding='utf-8')))}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
