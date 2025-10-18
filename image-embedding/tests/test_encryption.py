import pytest
import json
import os

from dotenv import load_dotenv
from image_embedding.encryption import Encrypter

load_dotenv()

skip_encryption = os.getenv("E2E_PRIVATE_KEY") is None


@pytest.fixture()
def encrypted_decrypted() -> tuple[dict, dict]:
    return {
        "value": "eyJqc29uX3BheWxvYWQiOiI3NEhjQ0dSSFljM0YyL1d6My85OWhqN1ZRa25VQkRiZVBKNDlEZ05ocTcrMXBaNGE1K1lVUWptMDM2ODhoTUEwdzBzUThjNjlkNXNLc1JMSE9aQW04SlMvUVRKaWU4RG1HSmtZcG5VZXU1cDRHYitqdXJOSlRCUU0iLCJhZXNfa2V5IjoiWCtoYWxmQlBNYlBNRXgyeXd1S0J6c2pxNFpFVG55Y3BNejl5b3MvWDRBOXFzR2tlaXRjZ01TTzFqU0xsK2J0M1VJVmFpdzlaekpSUjlRci9TV3Q3aWU5elhUTVg0R3V3eTVSMk9KbDZ5cjNYNmJLSWtGc2s0RFNORGlCeGd6N0hNYW9ldWVhN0tYclpTQzZrOTNxSXdTVHBUbnd0ejdxRGFxMWNCdHYyQUJaOXZPWmFqRmpMNk1JVXI0OWNSaFk5UFV0eExRVXA1bEJZNnRDT21zR21JNGRvRld2UGp0cVpkdzZqcm9YK2l1YmJMUFpiYVo0a3lOMXNmczg0ck1XVGptR0RmdVVqcDRYeFFxVk1mcnhReTJtTTdhQ3QyS2hEdjZSemIwWHdvL2sxOVJ3eWoycXhzcnNDUHFWQ0lEYjh3ZXNwVk5KOFhHTzcyYjNHN0hOZ2V3PT0ifQ=="
    }, {"image_url": "hello world", "api_key": "this is not an api key"}


@pytest.mark.skipif(
    condition=skip_encryption, reason="Environment setup misses E2E_PRIVATE_KEY"
)
def test_encrypt_decrypt(encrypted_decrypted: tuple[dict, dict]) -> None:
    enc = Encrypter()
    deciphered_message = json.loads(enc.decrypt(encrypted_decrypted[0]).decode("utf-8"))
    assert deciphered_message == encrypted_decrypted[1]
