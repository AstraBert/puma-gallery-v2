import os

from Crypto.PublicKey import RSA
from Crypto.Cipher import PKCS1_OAEP, AES
from Crypto.Hash import SHA256

class Encrypter:
    def __init__(self) -> None:
        self._private_key = RSA.import_key(os.getenv("E2E_PRIVATE_KEY", ""))
        self._public_key = self._private_key.public_key()
        
    def decrypt(self, aes_key: bytes, data: bytes) -> bytes:
        cipher_rsa = PKCS1_OAEP.new(self._private_key, hashAlgo=SHA256)
        key = cipher_rsa.decrypt(aes_key)

        nonce = data[:12]
        ciphertext_and_tag = data[12:]

        cipher_aes = AES.new(key, mode=AES.MODE_GCM, nonce=nonce)
        return cipher_aes.decrypt_and_verify(
            ciphertext_and_tag[:-16],
            ciphertext_and_tag[-16:]
        )