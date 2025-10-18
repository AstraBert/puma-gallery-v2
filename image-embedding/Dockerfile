FROM ghcr.io/astral-sh/uv:python3.13-alpine

WORKDIR /app/

COPY ./ /app/

RUN uv sync

ENV PYTHONUNBUFFERED=1
ENTRYPOINT ["/bin/sh", "-c", "source .venv/bin/activate && embed-images"]
