"""Structured logging setup for the quantum module."""

from __future__ import annotations

import logging
import sys


def configure_logging(level: str = "INFO") -> logging.Logger:
    """Configure root logger for CLI/library use and return the package logger."""
    numeric = getattr(logging, level.upper(), logging.INFO)
    root = logging.getLogger()
    if not root.handlers:
        handler = logging.StreamHandler(sys.stderr)
        handler.setFormatter(
            logging.Formatter("%(asctime)s %(levelname)s [%(name)s] %(message)s")
        )
        root.addHandler(handler)
    root.setLevel(numeric)
    # Keep third-party circuit passmanager noise off the CLI by default.
    if numeric > logging.DEBUG:
        logging.getLogger("qiskit").setLevel(logging.WARNING)
    logger = logging.getLogger("qguard_quantum")
    logger.setLevel(numeric)
    return logger
