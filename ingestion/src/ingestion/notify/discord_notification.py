import os

from httpx import AsyncClient, HTTPError

async def send_discord_notification(content: str) -> bool:
    data = {"content": content}
    async with AsyncClient() as client:
        try:
            response = await client.post(os.getenv("DISCORD_WEBHOOK_URL", ""), json=data)
            response.raise_for_status()
            return response.status_code == 204
        except HTTPError:
            return False