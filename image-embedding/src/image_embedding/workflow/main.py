import asyncio
import json
import os
import base64

from typing import Optional
from workflows import Workflow, step
from workflows.events import StartEvent, StopEvent
from kafka import KafkaConsumer, KafkaProducer
from image_embedding.embeddings import produce_embeddings
from image_embedding.encryption import Encrypter

class InputEvent(StartEvent):
    image_url: str

class OutputEvent(StopEvent):
    embeddings: Optional[list[float]] = None
    image_url: Optional[str] = None
    api_key: str
    error: Optional[str] = None

class EmbedImageWorkflow(Workflow):
    @step
    async def embed_image(self, ev: InputEvent) -> OutputEvent:
        embeddings = await produce_embeddings(image_url=ev.image_url)
        return OutputEvent(embeddings=embeddings, image_url=ev.image_url, api_key=os.getenv("KAFKA_API_KEY", ""))
    
async def run_workflow():
    await asyncio.sleep(30)
    enc = Encrypter()
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
    
    wf = EmbedImageWorkflow(timeout=600)
    
    print("Starting to consume messages from 'images' topic...")
    
    try:
        while True:
            # Poll for messages with timeout
            message_batch = consumer.poll(timeout_ms=1000)
            
            for _, messages in message_batch.items():
                for message in messages:
                    print(f"Received a message")
                    try:
                        deciphered_message = json.loads(enc.decrypt(message.value).decode("utf-8"))
                        if deciphered_message.get("api_key", "") != os.getenv("KAFKA_API_KEY", ""):
                            continue
                        output = await wf.run(start_event=InputEvent(image_url=deciphered_message. get("image_url", "")))
                        data, key = enc.encrypt(json.dumps(output.model_dump()).encode("utf-8"))
                        value = {
                            "aes_key": base64.b64encode(key).decode('utf-8'),
                            "json_payload": base64.b64encode(data).decode('utf-8')
                        }
                        producer.send(topic="image-embeddings", value=value)
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
    
