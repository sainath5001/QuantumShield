"""Input validation for educational demos."""

from __future__ import annotations

from typing import Iterable


class ValidationError(ValueError):
    """Raised when demo inputs are unsupported or invalid."""


def require_positive_int(name: str, value: int, *, maximum: int | None = None) -> int:
    if not isinstance(value, int) or isinstance(value, bool):
        raise ValidationError(f"{name} must be an int, got {type(value).__name__}")
    if value <= 0:
        raise ValidationError(f"{name} must be > 0, got {value}")
    if maximum is not None and value > maximum:
        raise ValidationError(f"{name} must be <= {maximum} for this educational demo, got {value}")
    return value


def require_bit_string(name: str, value: str, *, width: int) -> str:
    if not isinstance(value, str):
        raise ValidationError(f"{name} must be a string")
    if len(value) != width or any(c not in "01" for c in value):
        raise ValidationError(f"{name} must be a {width}-bit string of 0/1, got {value!r}")
    return value


def require_odd_composite(n: int, *, allowed: Iterable[int]) -> int:
    require_positive_int("n", n)
    allowed_set = set(allowed)
    if n not in allowed_set:
        raise ValidationError(
            f"Unsupported n={n}. Educational Shor demo only allows small integers "
            f"{sorted(allowed_set)}. This is not a secp256k1 attack."
        )
    return n


def require_shots(shots: int, *, maximum: int = 4096) -> int:
    return require_positive_int("shots", shots, maximum=maximum)


def require_seed(seed: int | None) -> int | None:
    if seed is None:
        return None
    if not isinstance(seed, int) or isinstance(seed, bool):
        raise ValidationError("seed must be an int or None")
    return seed
