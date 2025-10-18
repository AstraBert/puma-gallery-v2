import pytest
import os

from dotenv import load_dotenv
from ingestion.notify import send_discord_notification

load_dotenv()


@pytest.mark.asyncio
@pytest.mark.skipif(
    condition=os.getenv("DISCORD_WEBHOOK_URL") is None,
    reason="Environment setup misses DISCORD_WEBHOOK_URL",
)
async def test_send_discord_notification() -> None:
    assert await send_discord_notification("this is a test message")
