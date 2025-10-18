import pytest

from unittest.mock import AsyncMock, patch
from image_embedding.embeddings import produce_embeddings


@pytest.mark.asyncio
async def test_produce_embeddings() -> None:
    mock_embeddings = [0.1, 0.2, 0.3, 0.4, 0.5]
    with patch("replicate.async_run", new_callable=AsyncMock) as mock_async_run:
        mock_async_run.return_value = mock_embeddings
        result = await produce_embeddings("http://example.com/image.jpg")
        assert result == mock_embeddings
        mock_async_run.assert_called_once_with(
            "daanelson/imagebind:0383f62e173dc821ec52663ed22a076d9c970549c209666ac3db181618b7a304",
            input={"input": "http://example.com/image.jpg", "modality": "vision"},
        )
