import asyncio
import json

from typing import Union, Optional
from httpx import AsyncClient, HTTPError
from workflows import Workflow, step
from workflows.events import StartEvent, StopEvent, Event
from kafka import KafkaConsumer, KafkaProducer
from image_embedding.embeddings import produce_embeddings

class InputEvent(StartEvent):
    image_url: str

class DownloadedImageEvent(Event):
    image_bytes: bytes
    image_url: str

class OutputEvent(StopEvent):
    embeddings: Optional[list[float]] = None
    image_url: Optional[str] = None
    error: Optional[str] = None

class EmbedImageWorkflow(Workflow):
    @step
    async def download_image_from_url(self, ev: InputEvent) -> Union[DownloadedImageEvent, OutputEvent]:
        try:
            async with AsyncClient() as client:
                response = await client.get(ev.image_url)
                response.raise_for_status()
                content = response.content
                return DownloadedImageEvent(image_bytes=content, image_url=ev.image_url)
        except HTTPError as e:
            return OutputEvent(error=f"An error occurred while downloading the image: {e}")

    @step
    async def embed_image(self, ev: DownloadedImageEvent) -> OutputEvent:
        embeddings = produce_embeddings(image_content=ev.image_bytes)
        return OutputEvent(embeddings=embeddings, image_url=ev.image_url)
    
async def run_workflow():
    consumer = KafkaConsumer(
        'images',
        bootstrap_servers='kafka:9092',
        auto_offset_reset='earliest',
        enable_auto_commit=True, 
        group_id='images-consumer-0',
        value_deserializer=lambda x: json.loads(x.decode('utf-8')),
        consumer_timeout_ms=1000
    )
    
    producer = KafkaProducer(
        bootstrap_servers=['kafka:9092'],
        value_serializer=lambda v: json.dumps(v).encode('utf-8')
    )
    
    wf = EmbedImageWorkflow(timeout=120)
    
    print("Starting to consume messages from 'images' topic...")
    
    try:
        while True:
            # Poll for messages with timeout
            message_batch = consumer.poll(timeout_ms=1000)
            
            for _, messages in message_batch.items():
                for message in messages:
                    print(f"Received message: {message.value}")
                    try:
                        output = await wf.run(start_event=InputEvent(image_url=message.value.get("image_url", "")))
                        producer.send(topic="image-embeddings", value=output.model_dump())
                        producer.flush()  # Ensure message is sent
                    except Exception as e:
                        print(f"Error processing message: {e}")
            
            # Small sleep to prevent CPU spinning
            await asyncio.sleep(0.1)
            
    except KeyboardInterrupt:
        print("Shutting down consumer...")
    finally:
        consumer.close()
        producer.close()

def main():
    asyncio.run(run_workflow())
    
