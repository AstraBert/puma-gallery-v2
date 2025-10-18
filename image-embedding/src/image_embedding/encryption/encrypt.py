import os
import base64
import json

from Crypto.PublicKey import RSA
from Crypto.Cipher import PKCS1_OAEP, AES
from Crypto.Hash import SHA256


def load_private_key_from_env() -> str:
    b64_enc = os.getenv("E2E_PRIVATE_KEY", "")
    return base64.b64decode(b64_enc).decode("utf-8")


class Encrypter:
    def __init__(self) -> None:
        self._private_key = RSA.import_key(load_private_key_from_env())
        self._public_key = self._private_key.public_key()

    def decrypt(self, message: dict) -> bytes:
        b64_data = base64.b64decode(message.get("value", ""))
        d = json.loads(b64_data.decode("utf-8"))
        cipher_rsa = PKCS1_OAEP.new(self._private_key, hashAlgo=SHA256)
        key = cipher_rsa.decrypt(base64.b64decode(d.get("aes_key")))
        data = base64.b64decode(d.get("json_payload"))
        nonce = data[:12]
        ciphertext_and_tag = data[12:]
        cipher_aes = AES.new(key, mode=AES.MODE_GCM, nonce=nonce)
        return cipher_aes.decrypt_and_verify(
            ciphertext_and_tag[:-16], ciphertext_and_tag[-16:]
        )

    def encrypt(self, data: bytes) -> tuple[bytes, bytes]:
        new_aes_key = os.urandom(32)
        nonce = os.urandom(12)  # explicitly 12 bytes
        cipher_aes = AES.new(new_aes_key, mode=AES.MODE_GCM, nonce=nonce)
        ciphertext, tag = cipher_aes.encrypt_and_digest(data)
        encrypted_data = nonce + ciphertext + tag
        cipher_rsa = PKCS1_OAEP.new(self._public_key, hashAlgo=SHA256)
        encrypted_new_key = cipher_rsa.encrypt(new_aes_key)
        return encrypted_data, encrypted_new_key
