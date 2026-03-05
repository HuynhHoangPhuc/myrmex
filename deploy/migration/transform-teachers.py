#!/usr/bin/env python3
"""
Transform HCMUS teacher spreadsheet → teachers.csv for bulk import API.
Usage: python3 deploy/migration/transform-teachers.py --input hcmus-teachers.xlsx --output data/teachers.csv
"""
import argparse
from pathlib import Path
import pandas as pd


COLUMN_MAP = {
    # HCMUS column name → import format column name
    "Họ và tên":       "full_name",
    "Email":           "email",
    "Mã giảng viên":   "employee_code",
    "Khoa":            "department_name",
    "Số điện thoại":   "phone",
    "Chuyên môn":      "specializations",
    "Số tiết tối đa":  "max_hours_per_week",
}

REQUIRED_OUTPUT_COLS = ["full_name", "email", "employee_code", "department_name"]
DEFAULT_MAX_HOURS = 12


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input",  required=True, help="HCMUS teacher spreadsheet (.xlsx or .csv)")
    parser.add_argument("--output", default="data/teachers.csv")
    args = parser.parse_args()

    src = Path(args.input)
    df = pd.read_excel(src) if src.suffix in (".xlsx", ".xls") else pd.read_csv(src, encoding="utf-8-sig")

    # Rename known columns, keep unknowns as-is
    df = df.rename(columns={k: v for k, v in COLUMN_MAP.items() if k in df.columns})

    # Ensure required columns exist
    for col in REQUIRED_OUTPUT_COLS:
        if col not in df.columns:
            raise ValueError(f"Required column missing after mapping: '{col}'. Check COLUMN_MAP.")

    # Normalise
    df["email"] = df["email"].str.strip().str.lower()
    df["full_name"] = df["full_name"].str.strip()
    if "max_hours_per_week" not in df.columns:
        df["max_hours_per_week"] = DEFAULT_MAX_HOURS
    else:
        df["max_hours_per_week"] = df["max_hours_per_week"].fillna(DEFAULT_MAX_HOURS).astype(int)

    # Drop rows with missing required fields
    before = len(df)
    df = df.dropna(subset=REQUIRED_OUTPUT_COLS)
    dropped = before - len(df)
    if dropped:
        print(f"WARNING: dropped {dropped} rows with missing required fields")

    out = Path(args.output)
    out.parent.mkdir(parents=True, exist_ok=True)
    df.to_csv(out, index=False, encoding="utf-8")
    print(f"✅ Exported {len(df)} teachers → {out}")


if __name__ == "__main__":
    main()
