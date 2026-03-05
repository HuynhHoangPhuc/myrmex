#!/usr/bin/env python3
"""
Pre-flight validation for HCMUS migration data.
Checks structure, encoding, duplicates, and cross-references before import.

Usage:
    python3 deploy/migration/validate-data.py --input data/ --report validation-report.md
"""
import argparse
import sys
from pathlib import Path

import pandas as pd


def load_csv(path: Path, required_cols: list[str]) -> pd.DataFrame | None:
    if not path.exists():
        print(f"  SKIP: {path} not found")
        return None
    df = pd.read_csv(path, encoding="utf-8-sig")  # utf-8-sig handles BOM from Excel exports
    missing = [c for c in required_cols if c not in df.columns]
    if missing:
        print(f"  FAIL: {path.name} missing columns: {missing}")
        return None
    return df


def check_duplicates(df: pd.DataFrame, col: str, label: str) -> list[str]:
    dupes = df[df[col].duplicated(keep=False)][col].dropna().unique().tolist()
    if dupes:
        return [f"Duplicate {label}: {v}" for v in dupes[:10]]
    return []


def validate(input_dir: Path) -> tuple[list[str], list[str]]:
    errors: list[str] = []
    warnings: list[str] = []

    teachers_path  = input_dir / "teachers.csv"
    students_path  = input_dir / "students.csv"
    depts_path     = input_dir / "departments.json"

    # --- Teachers ---
    print("Validating teachers.csv...")
    teachers_cols = ["full_name", "email", "employee_code", "department_name"]
    teachers = load_csv(teachers_path, teachers_cols)
    if teachers is not None:
        errors += check_duplicates(teachers, "email", "teacher email")
        errors += check_duplicates(teachers, "employee_code", "employee_code")
        null_emails = teachers["email"].isna().sum()
        if null_emails:
            errors.append(f"Teachers: {null_emails} rows with missing email")
        # Check encoding (Vietnamese characters)
        try:
            teachers["full_name"].str.encode("utf-8")
        except Exception as e:
            errors.append(f"Teachers: encoding issue in full_name — {e}")
        print(f"  {len(teachers)} teachers loaded")

    # --- Students ---
    print("Validating students.csv...")
    students_cols = ["full_name", "email", "student_code", "department_name"]
    students = load_csv(students_path, students_cols)
    if students is not None:
        errors += check_duplicates(students, "email", "student email")
        errors += check_duplicates(students, "student_code", "student_code")
        null_emails = students["email"].isna().sum()
        if null_emails:
            errors.append(f"Students: {null_emails} rows with missing email")
        print(f"  {len(students)} students loaded")

    # --- Cross-check: no email overlap between teachers and students ---
    if teachers is not None and students is not None:
        teacher_emails = set(teachers["email"].dropna().str.lower())
        student_emails = set(students["email"].dropna().str.lower())
        overlap = teacher_emails & student_emails
        if overlap:
            for email in list(overlap)[:5]:
                errors.append(f"Email appears in both teachers and students: {email}")

    # --- Departments ---
    print("Validating departments.json...")
    if depts_path.exists():
        import json
        depts = json.loads(depts_path.read_text())
        dept_names = {d["name"] for d in depts}

        if teachers is not None:
            unknown_depts = set(teachers["department_name"].dropna()) - dept_names
            for d in unknown_depts:
                errors.append(f"Teachers: unknown department_name '{d}'")

        if students is not None:
            unknown_depts = set(students["department_name"].dropna()) - dept_names
            for d in unknown_depts:
                errors.append(f"Students: unknown department_name '{d}'")

        print(f"  {len(depts)} departments loaded")
    else:
        warnings.append("departments.json not found — department cross-check skipped")

    return errors, warnings


def main() -> None:
    parser = argparse.ArgumentParser(description="Pre-flight validation for HCMUS migration data")
    parser.add_argument("--input",  required=True, help="Directory containing CSV/JSON source files")
    parser.add_argument("--report", default="validation-report.md", help="Output report file path")
    args = parser.parse_args()

    input_dir = Path(args.input)
    if not input_dir.is_dir():
        print(f"ERROR: {input_dir} is not a directory")
        sys.exit(1)

    print(f"\n=== Pre-flight validation: {input_dir} ===\n")
    errors, warnings = validate(input_dir)

    report_lines = [
        "# Migration Pre-flight Validation Report\n",
        f"**Input**: `{input_dir}`\n",
        f"**Result**: {'✅ PASS' if not errors else '❌ FAIL'}\n",
        f"**Errors**: {len(errors)} | **Warnings**: {len(warnings)}\n",
    ]

    if errors:
        report_lines += ["\n## Errors\n"] + [f"- {e}\n" for e in errors]
    if warnings:
        report_lines += ["\n## Warnings\n"] + [f"- {w}\n" for w in warnings]

    Path(args.report).write_text("".join(report_lines))
    print(f"\nReport written to {args.report}")

    if errors:
        print(f"\n❌ VALIDATION FAILED — {len(errors)} error(s). Fix before importing.")
        sys.exit(1)
    else:
        print(f"\n✅ VALIDATION PASSED{' (with warnings)' if warnings else ''}. Ready to import.")


if __name__ == "__main__":
    main()
