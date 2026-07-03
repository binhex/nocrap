#!/usr/bin/env python3
"""Run pytest --crap on the corpus and output reference CRAP scores as JSON.

Usage: python3 reference.py <corpus_dir>

This uses the exact pytest --crap workflow from the tech-debt skill.
Output: JSON array of {name, file, cc, coverage_percent, crap} objects.
"""

import json
import os
import re
import subprocess
import sys


def run_pytest_crap(corpus_dir: str) -> str:
    """Run pytest --crap on the corpus directory and return stdout."""
    cmd = [
        sys.executable, "-m", "pytest",
        "--cov=" + corpus_dir,
        "--cov-report=json",
        "--crap", "--crap-top-n=0",
        "--tb=no", "-q",
        corpus_dir,
    ]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode not in (0, 1):
        print(result.stderr, file=sys.stderr)
        print("stdout:", result.stdout[:500], file=sys.stderr)
        result.check_returncode()
    return result.stdout


def parse_crap_table(output: str) -> list[dict]:
    """Parse the CRAP by Function table from pytest --crap output.

    The Rich table uses box-drawing characters:
      ┏━━━━━━━━━┳━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┓
      ┃    CRAP ┃  CC ┃   Coverage ┃ Function          ┃ File                 ┃
      ┡━━━━━━━━━╇━━━━━╇━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━┩
      │   14.00 │  11 │      70.8% │ all_branches      │ corpus_py/branches.py│
    """
    scores = []
    in_table = False
    ansi_re = re.compile(r'\x1b\[[0-9;]*m')
    sep_re = re.compile(r'[│┃]')

    for line in output.split("\n"):
        stripped = line.strip()
        if "CRAP by Function" in stripped:
            in_table = True
            continue
        if not in_table or not stripped:
            continue
        if stripped.startswith("┏") or stripped.startswith("┃") and stripped.startswith("CRAP"):
            continue
        if stripped.startswith("┡") or stripped.startswith("└"):
            continue
        if stripped.startswith("┃") and "CRAP" in stripped:
            continue

        # Strip ANSI codes
        clean = ansi_re.sub("", stripped)

        # Handle Rich table row: │ value1 │ value2 │ value3 │ value4 │ value5 │
        if clean.startswith("│"):
            parts = sep_re.split(clean)
            parts = [p.strip() for p in parts if p.strip()]
            if len(parts) >= 5:
                try:
                    crap = float(parts[0])
                    cc = int(parts[1])
                    cov_str = parts[2].rstrip("%")
                    cov_pct = float(cov_str)
                    name = parts[3]
                    filepath = parts[4]
                    scores.append({
                        "name": name,
                        "file": filepath,
                        "cc": cc,
                        "coverage_percent": cov_pct,
                        "crap": crap,
                    })
                except (ValueError, IndexError):
                    continue

    return scores


def main():
    corpus_dir = sys.argv[1] if len(sys.argv) > 1 else "corpus_py"
    output = run_pytest_crap(corpus_dir)
    scores = parse_crap_table(output)

    scores.sort(key=lambda s: (s["file"], s["name"]))

    print(json.dumps(scores, indent=2))

    # Also save for debugging
    with open("reference_scores.json", "w") as f:
        json.dump(scores, f, indent=2)


if __name__ == "__main__":
    main()
