import asyncio
import json
import os
import base64

from typing import Optional
from workflows import Workflow, step
from workflows.events import StartEvent, StopEvent
from kafka import KafkaConsumer
from ingestion.vector_store import ImageUploader
from ingestion.notify import send_discord_notification
from ingestion.encryption import Decrypter


class InputEvent(StartEvent):
    image_url: str
    image_embeddings: list[float]


class OutputEvent(StopEvent):
    success: bool
    error: Optional[str] = None


class EmbedImageWorkflow(Workflow):
    @step
    async def upload_image(self, ev: InputEvent) -> OutputEvent:
        uploader = ImageUploader(
            qdrant_url=os.getenv("QDRANT_URL", ""),
            qdrant_api_key=os.getenv("QDRANT_API_KEY", ""),
            collection_name=os.getenv("QDRANT_COLLECTION_NAME", ""),
        )
        succ = await uploader.upload_image(ev.image_embeddings, ev.image_url)
        error: Optional[str] = None
        if succ:
            notf_sent = await send_discord_notification(
                f"Successfully embedded and uploaded this image:\n\n![pumito the catito]({ev.image_url})"
            )
        else:
            error = "An error occurred while uploading the image to Qdrant"
            notf_sent = await send_discord_notification(
                f"It was not possible to embedd and upload this image:\n\n![pumito the catito]({ev.image_url})"
            )
        if not notf_sent and not error:
            error = "An error occurred while sending the notification to Discord"
        elif not notf_sent and error is not None:
            error += "An error occurred while sending the notification to Discord"
        return OutputEvent(success=(succ and notf_sent), error=error)


async def run_workflow() -> None:
    await asyncio.sleep(30)
    enc = Decrypter()
    consumer = KafkaConsumer(
        "image-embeddings",
        bootstrap_servers="kafka:9092",
        auto_offset_reset="earliest",
        enable_auto_commit=True,
        group_id="emebddings-consumer-0",
        value_deserializer=lambda x: json.loads((x).decode("utf-8")),
        consumer_timeout_ms=1000,
    )

    wf = EmbedImageWorkflow(timeout=300)

    print("Starting to consume messages from 'image-embeddings' topic...")

    try:
        while True:
            # Poll for messages with timeout
            message_batch = consumer.poll(timeout_ms=1000)

            for _, messages in message_batch.items():
                for message in messages:
                    print("Received message")
                    aes_key = base64.b64decode(message.value.get("aes_key", ""))
                    json_payload = base64.b64decode(
                        message.value.get("json_payload", "")
                    )
                    deciphered_message = json.loads(enc.decrypt(aes_key, json_payload))
                    if deciphered_message.get("api_key", "") != os.getenv(
                        "KAFKA_API_KEY", ""
                    ):
                        continue
                    try:
                        await wf.run(
                            start_event=InputEvent(
                                image_url=deciphered_message.get("image_url", ""),
                                image_embeddings=deciphered_message.get(
                                    "image_embeddings", []
                                ),
                            )
                        )
                    except Exception as e:
                        print(f"Error processing message: {e}")

            # Small sleep to prevent CPU spinning
            await asyncio.sleep(0.1)

    except KeyboardInterrupt:
        print("Shutting down consumer...")
    finally:
        consumer.close()


def main() -> None:
    asyncio.run(run_workflow())
