import replicate

from typing import cast

async def produce_embeddings(image_url: str) -> list[float]:
    prediction = await replicate.async_run(
        "daanelson/imagebind:0383f62e173dc821ec52663ed22a076d9c970549c209666ac3db181618b7a304",
        input={
            "input": image_url,
            "modality": "vision"
        }
    )
    return cast(list[float], prediction)