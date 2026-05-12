SYSTEM_PROMPT = """You are QueryForge SQL Generator.
You generate safe SQLite SELECT queries from natural language.
You must only output valid JSON.
Never output markdown.
Never output explanations outside JSON.
Only generate read-only SQL.
Allowed statements:
- SELECT
- WITH clauses that only contain SELECT
Forbidden:
- INSERT
- UPDATE
- DELETE
- DROP
- ALTER
- CREATE
- TRUNCATE
- ATTACH
- DETACH
- PRAGMA
- VACUUM
- REINDEX
- multiple statements
Return JSON exactly in this format:
{
  "sql": "SELECT ...",
  "explanation": "short explanation",
  "confidence": 0.0
}
"""


def build_schema_text(schema: dict) -> str:
    lines = []
    for table in schema.get("tables", []):
        columns = ", ".join(f"{c['name']} {c.get('type') or 'TEXT'}" for c in table.get("columns") or [])
        lines.append(f"- {table['name']}({columns})")
        for fk in table.get("foreign_keys") or []:
            lines.append(f"  FK {table['name']}.{fk['column']} -> {fk['ref_table']}.{fk['ref_column']}")
    return "\n".join(lines)


def build_user_prompt(question: str, schema: dict, safety_rules: list[str]) -> str:
    return (
        "Database schema:\n"
        f"{build_schema_text(schema)}\n\n"
        f"User question:\n{question}\n\n"
        "Use SQLite syntax. Include a LIMIT unless the query is an aggregation where LIMIT is unnecessary. "
        "Use only the schema above. Follow these backend safety rules too:\n"
        + "\n".join(f"- {rule}" for rule in safety_rules)
    )
