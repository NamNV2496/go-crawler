from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, root_validator
import requests
import os
from dotenv import load_dotenv
import uvicorn

# Load environment variables
load_dotenv()

app = FastAPI()
LLM_API_URL = os.getenv("LLM_API_URL")

ALLOWED_ROLES = {"system", "user", "assistant"}
MAX_MESSAGES = 20
MAX_CONTENT_LENGTH = 2000

class Message(BaseModel):
    role: str
    content: str

class ChatRequest(BaseModel):
    messages: list[Message]
    model: str = "mistral-7b-instruct-v0.3"
    temperature: float = 0.7
    max_tokens: int = 500

@app.post("/chat")
def chat(request: ChatRequest):
    payload = {
        "model": request.model,
        "messages": [msg.dict() for msg in request.messages],
        "temperature": request.temperature,
        "max_tokens": request.max_tokens,
        "stream": False
    }
    try:
        response = requests.post(LLM_API_URL, json=payload, timeout=60)
        response.raise_for_status()
        data = response.json()
        # Return more info (keep all, or filter keys as you want)
        return {
            "choices": data.get("choices", []),
            "usage": data.get("usage"),
            "model": data.get("model"),
            "created": data.get("created"),
            "response": data["choices"][0]["message"]["content"]
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    
    uvicorn.run("main:app", host="0.0.0.1", port=os.getenv("PORT"), reload=True)

