#!/usr/bin/env python3
"""
Transform HCMUS student spreadsheet → students.csv for bulk import API.
Usage: python3 deploy/migration/transform-students.py --input hcmus-students.xlsx --output data/students.csv
"""
import argparse
from pathlib import Path
import pandas as pd


COLUMN_MAP = {
    "Họ và tên":  "full_name",
    "Email":      "email",
    "MSSV":       "student_code",
    "Khoa":       "department_name",
}

REQUIRED_OUTPUT_COLS = ["full_name", "email", "student_code", "department_name"]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input",  required=True, help="HCMUS student spreadsheet (.xlsx or .csv)")
    parser.add_argument("--output", default="data/students.csv")
    args = parser.parse_args()

    src = Path(args.input)
    df = pd.read_excel(src) if src.suffix in (".xlsx", ".xls") else pd.read_csv(src, encoding="utf-8-sig")

    df = df.rename(columns={k: v for k, v in COLUMN_MAP.items() if k in df.columns})

    for col in REQUIRED_OUTPUT_COLS:
        if col not in df.columns:
            raise ValueError(f"Required column missing after mapping: '{col}'. Check COLUMN_MAP.")

    df["email"] = df["email"].str.strip().str.lower()
    df["full_name"] = df["full_name"].str.strip()
    df["student_code"] = df["student_code"].astype(str).str.strip()

    before = len(df)
    df = df.dropna(subset=REQUIRED_OUTPUT_COLS)
    dropped = before - len(df)
    if dropped:
        print(f"WARNING: dropped {dropped} rows with missing required fields")

    out = Path(args.output)
    out.parent.mkdir(parents=True, exist_ok=True)
    df.to_csv(out, index=False, encoding="utf-8")
    print(f"✅ Exported {len(df)} students → {out}")


if __name__ == "__main__":
    main()
