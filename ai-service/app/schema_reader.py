from typing import Any


def first_table(schema: dict[str, Any]) -> dict[str, Any] | None:
    tables = schema.get("tables", [])
    return tables[0] if tables else None


def find_table(schema: dict[str, Any], question: str) -> dict[str, Any] | None:
    question_l = question.lower()
    tables = schema.get("tables", [])
    for table in tables:
        name = table["name"].lower()
        singular = name[:-1] if name.endswith("s") else name
        if name in question_l or singular in question_l:
            return table
    return first_table(schema)


def column_names(table: dict[str, Any]) -> list[str]:
    return [column["name"] for column in table.get("columns", [])]


def numeric_columns(table: dict[str, Any]) -> list[str]:
    numeric_markers = ("INT", "REAL", "NUM", "DEC", "DOUBLE", "FLOAT")
    return [
        column["name"]
        for column in table.get("columns", [])
        if any(marker in (column.get("type") or "").upper() for marker in numeric_markers)
    ]
