from uuid import uuid4
from qdrant_client.async_qdrant_client import AsyncQdrantClient
from qdrant_client.models import PointStruct

class ImageUploader:
    def __init__(self, qdrant_url: str, qdrant_api_key: str, collection_name: str, threshold: float = 0.75) -> None:
        self.qdrant_url = qdrant_url
        self.qdrant_api_key = qdrant_api_key
        self.collection_name = collection_name
        self.threshold = threshold
    
    def _get_qdrant_client(self):
        return AsyncQdrantClient(url=self.qdrant_url, api_key=self.qdrant_api_key)
    
    async def upload_image(self, image_embedding: list[float], image_url: str) -> bool:
        client = self._get_qdrant_client()
        point = PointStruct(id=str(uuid4()), vector=image_embedding, payload={"image_url": image_url})
        try:
            client.upload_points(collection_name=self.collection_name, points=[point])
            return True
        except Exception:
            return False


