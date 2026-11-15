"""Q-Guard educational quantum research module.

This package demonstrates quantum algorithms relevant to cryptography education.
It must never be connected to Ethereum private keys and never claims to attack secp256k1.
"""

from __future__ import annotations

__version__ = "0.1.0"

LIMITATIONS_BANNER = (
    "Educational / research demonstration only. "
    "Not a practical attack on Ethereum secp256k1 keys. "
    "Results must never be used as production wallet-scan evidence."
)

__all__ = [
    "__version__",
    "LIMITATIONS_BANNER",
]
