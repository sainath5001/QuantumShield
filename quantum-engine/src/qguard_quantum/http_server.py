"""ASGI entrypoint for the educational quantum FastAPI sidecar."""

from __future__ import annotations

import os


def main() -> None:
    import uvicorn

    host = os.environ.get("QGUARD_QUANTUM_HOST", "127.0.0.1")
    port = int(os.environ.get("QGUARD_QUANTUM_PORT", "8090"))
    uvicorn.run(
        "qguard_quantum.http_api:app",
        host=host,
        port=port,
        reload=False,
        log_level=os.environ.get("QGUARD_QUANTUM_LOG_LEVEL", "info"),
    )


if __name__ == "__main__":
    main()
