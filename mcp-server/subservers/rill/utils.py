def prune(obj):
    """
    Recursively remove keys with empty, null, or non-substantial values from dicts/lists.
    """
    if isinstance(obj, dict):
        return {
            k: prune(v)
            for k, v in obj.items()
            if v not in (None, "", [], {})
            and not (isinstance(v, dict) and not v)
            and not (isinstance(v, list) and not v)
        }
    elif isinstance(obj, list):
        return [
            prune(v)
            for v in obj
            if v not in (None, "", [], {})
            and not (isinstance(v, dict) and not v)
            and not (isinstance(v, list) and not v)
        ]
    else:
        return obj
