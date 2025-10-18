import pytest

from typing_extensions import override
from qdrant_client.models import PointStruct
from ingestion.vector_store import ImageUploader


class MockQdrantClient:
    def __init__(self, url: str, api_key: str) -> None:
        pass

    def upload_points(self, collection_name: str, points: list[PointStruct]) -> None:
        if len(points) == 0:
            raise ValueError("a mock error")
        return None


class MockImageUploader(ImageUploader):
    @override
    def _get_qdrant_client(self) -> MockQdrantClient:
        return MockQdrantClient(self.qdrant_url, self.qdrant_api_key)


@pytest.mark.asyncio
async def test_image_uploader() -> None:
    img_up = MockImageUploader("", "", "")
    assert await img_up.upload_image([1, 2, 3, 4], "")
