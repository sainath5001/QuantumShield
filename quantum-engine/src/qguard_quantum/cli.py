"""Command-line interface for educational quantum demos."""

from __future__ import annotations

import argparse
import json
import sys
from typing import Any

from qguard_quantum.logging_config import configure_logging
from qguard_quantum.validation import ValidationError


_BULKY_CIRCUIT_KEYS = ("qasm", "draw_text")


def _compact_payload(payload: Any) -> Any:
    """Drop bulky circuit diagram/QASM fields for readable CLI output."""
    if isinstance(payload, dict):
        out: dict[str, Any] = {}
        for key, value in payload.items():
            if key in _BULKY_CIRCUIT_KEYS:
                continue
            out[key] = _compact_payload(value)
        return out
    if isinstance(payload, list):
        return [_compact_payload(item) for item in payload]
    return payload


def _print_json(payload: dict[str, Any], *, full: bool = False) -> None:
    data = payload if full else _compact_payload(payload)
    print(json.dumps(data, indent=2, sort_keys=False))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="qguard-quantum",
        description=(
            "Q-Guard educational quantum module (Qiskit). "
            "Toy demos only — not an Ethereum / secp256k1 attack tool."
        ),
    )
    parser.add_argument("-v", "--verbose", action="store_true", help="Enable debug logging")
    parser.add_argument(
        "--full",
        action="store_true",
        help="Include full circuit QASM and text diagrams (very large for Shor)",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    sub.add_parser("info", help="Module metadata and limitations")

    p_fund = sub.add_parser("fundamentals", help="Print qubit/gate educational explanations")
    p_fund.add_argument(
        "--demo",
        choices=("explain", "superposition", "bell", "gate"),
        default="explain",
    )
    p_fund.add_argument("--gate", default="h", help="Gate for gate demo (h|x|z)")
    p_fund.add_argument("--shots", type=int, default=512)
    p_fund.add_argument("--seed", type=int, default=42)

    p_shor = sub.add_parser("shor", help="Toy Shor-style factoring demo (N=15 only)")
    p_shor.add_argument("--n", type=int, default=15)
    p_shor.add_argument("--a", type=int, default=7)
    p_shor.add_argument("--counting-qubits", type=int, default=8, choices=(4, 8))
    p_shor.add_argument("--shots", type=int, default=1024)
    p_shor.add_argument("--seed", type=int, default=42)

    p_grover = sub.add_parser("grover", help="Tiny Grover search demo")
    p_grover.add_argument("--qubits", type=int, default=2)
    p_grover.add_argument("--target", default="11")
    p_grover.add_argument("--iterations", type=int, default=None)
    p_grover.add_argument("--shots", type=int, default=1024)
    p_grover.add_argument("--seed", type=int, default=42)

    p_pqc = sub.add_parser("pqc", help="Post-quantum cryptography education")
    p_pqc.add_argument("--algorithm", default=None, help="Optional: ML-DSA | ML-KEM | SLH-DSA")

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    configure_logging("DEBUG" if args.verbose else "INFO")

    full = bool(args.full)

    try:
        if args.command == "info":
            from qguard_quantum.interface import module_info

            _print_json(module_info(), full=full)
            return 0

        if args.command == "fundamentals":
            from qguard_quantum.interface import run_experiment

            if args.demo == "explain":
                _print_json(run_experiment("fundamentals"), full=full)
            elif args.demo == "superposition":
                _print_json(
                    run_experiment("superposition", shots=args.shots, seed=args.seed),
                    full=full,
                )
            elif args.demo == "bell":
                _print_json(
                    run_experiment("bell", shots=args.shots, seed=args.seed),
                    full=full,
                )
            else:
                _print_json(run_experiment("gate", gate=args.gate), full=full)
            return 0

        if args.command == "shor":
            from qguard_quantum.interface import run_experiment

            _print_json(
                run_experiment(
                    "shor",
                    n=args.n,
                    a=args.a,
                    counting_qubits=args.counting_qubits,
                    shots=args.shots,
                    seed=args.seed,
                ),
                full=full,
            )
            return 0

        if args.command == "grover":
            from qguard_quantum.interface import run_experiment

            kwargs: dict[str, Any] = {
                "n_qubits": args.qubits,
                "target": args.target,
                "shots": args.shots,
                "seed": args.seed,
            }
            if args.iterations is not None:
                kwargs["iterations"] = args.iterations
            _print_json(run_experiment("grover", **kwargs), full=full)
            return 0

        if args.command == "pqc":
            from qguard_quantum.interface import run_experiment

            if args.algorithm:
                _print_json(run_experiment("pqc-algo", name=args.algorithm), full=full)
            else:
                _print_json(run_experiment("pqc"), full=full)
            return 0

        parser.error(f"unhandled command {args.command}")
        return 2
    except ValidationError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    except Exception as exc:  # noqa: BLE001 — CLI boundary
        print(f"error: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
