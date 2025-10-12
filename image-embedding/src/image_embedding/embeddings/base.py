from typing import TYPE_CHECKING
from torch import device, no_grad
from PIL import Image
from io import BytesIO
from transformers.models.auto import AutoImageProcessor, AutoModel

DEVICE = device("cpu")
PROCESSOR = AutoImageProcessor.from_pretrained(pretrained_model_name_or_path="./model/preprocessor_config.json")
MODEL = AutoModel.from_pretrained(pretrained_model_name_or_path="./model/")

def produce_embeddings(image_content: bytes) -> list[float]:
    image = Image.open(BytesIO(image_content))
    inputs = PROCESSOR(images=image, return_tensors="pt").to(DEVICE)
    outputs = MODEL(**inputs)
    with no_grad():
        embeddings = outputs.last_hidden_state.mean(dim=1).cpu().numpy()
    return embeddings.tolist()[0]