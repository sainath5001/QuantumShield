"""Input validation tests."""

from __future__ import annotations

import pytest

from qguard_quantum.validation import (
    ValidationError,
    require_bit_string,
    require_odd_composite,
    require_positive_int,
    require_seed,
    require_shots,
)


def test_require_positive_int_ok() -> None:
    assert require_positive_int("x", 3, maximum=10) == 3


def test_require_positive_int_rejects_zero_and_oversize() -> None:
    with pytest.raises(ValidationError):
        require_positive_int("x", 0)
    with pytest.raises(ValidationError):
        require_positive_int("x", 11, maximum=10)
    with pytest.raises(ValidationError):
        require_positive_int("x", True)  # type: ignore[arg-type]


def test_require_bit_string() -> None:
    assert require_bit_string("t", "01", width=2) == "01"
    with pytest.raises(ValidationError):
        require_bit_string("t", "012", width=3)
    with pytest.raises(ValidationError):
        require_bit_string("t", "1", width=2)


def test_require_odd_composite_supported_only() -> None:
    assert require_odd_composite(15, allowed={15}) == 15
    with pytest.raises(ValidationError, match="secp256k1"):
        require_odd_composite(21, allowed={15})


def test_shots_and_seed() -> None:
    assert require_shots(100) == 100
    with pytest.raises(ValidationError):
        require_shots(10_000)
    assert require_seed(None) is None
    assert require_seed(7) == 7
    with pytest.raises(ValidationError):
        require_seed(1.5)  # type: ignore[arg-type]
