"""Package the ChatGPT plugin from a dcr-mcp release tag."""

import argparse
import json
import re
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile, ZipInfo


SOURCE = Path(__file__).resolve().parent
FILES = (
    "plugin.json",
    "mcp.json",
    "assets/dicty-icon.png",
    "skills/dicty-stock-catalog/SKILL.md",
)
TAG = re.compile(
    r"v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)"
    r"(?:-(?P<prerelease>[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?"
    r"(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?"
)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("tag", help="Git release tag, such as v0.1.4")
    parser.add_argument("output_dir", type=Path)
    args = parser.parse_args()
    match = TAG.fullmatch(args.tag)
    if match is None:
        parser.error("tag must be a v-prefixed semantic version")
    if re.search(r"(?:^|\.)0[0-9]+(?:\.|$)", match["prerelease"] or ""):
        parser.error("numeric prerelease identifiers cannot have leading zeros")

    contents = {name: (SOURCE / name).read_bytes() for name in FILES}
    manifest = json.loads(contents["plugin.json"])
    manifest["version"] = args.tag[1:]
    manifest["extensions"]["com.openai"]["publication"]["release_notes"] = (
        f"Packages the Dicty Stock Center catalog skill and production MCP "
        f"connection from dcr-mcp {args.tag}."
    )
    contents["plugin.json"] = (json.dumps(manifest, indent=2) + "\n").encode()

    args.output_dir.mkdir(parents=True, exist_ok=True)
    path = args.output_dir / f"{manifest['name']}-{args.tag}.zip"
    with ZipFile(path, "x") as archive:
        for name, content in contents.items():
            archive.writestr(ZipInfo(name), content, compress_type=ZIP_DEFLATED)
    print(path)


if __name__ == "__main__":
    main()
