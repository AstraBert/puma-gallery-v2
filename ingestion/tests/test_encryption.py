import os
import pytest

from dotenv import load_dotenv
from Crypto.Cipher import PKCS1_OAEP, AES
from Crypto.Hash import SHA256
from ingestion.encryption import Decrypter

load_dotenv()

skip_encryption = os.getenv("E2E_PRIVATE_KEY") is None


@pytest.mark.skipif(
    condition=skip_encryption, reason="Environment setup misses E2E_PRIVATE_KEY"
)
def test_decryption() -> None:
    dec = Decrypter()
    new_aes_key = os.urandom(32)
    nonce = os.urandom(12)
    cipher_aes = AES.new(new_aes_key, mode=AES.MODE_GCM, nonce=nonce)
    ciphertext, tag = cipher_aes.encrypt_and_digest(b"hello world")
    encrypted_data = nonce + ciphertext + tag
    cipher_rsa = PKCS1_OAEP.new(dec._public_key, hashAlgo=SHA256)
    encrypted_new_key = cipher_rsa.encrypt(new_aes_key)
    res = dec.decrypt(encrypted_new_key, encrypted_data)
    assert res == b"hello world"
