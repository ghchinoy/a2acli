# Agent Card Specification & Discoverability Guide

The **Agent Card** is the single public discovery document published by every A2A-compliant server at `/.well-known/agent-card.json`.

---

## 1. Agent Card Canonical JSON Structure (v1.0)

```json
{
  "name": "Document Summarizer Agent",
  "version": "1.0",
  "description": "Exposes document analysis and multi-format summarization capabilities.",
  "documentationUrl": "https://example.com/docs",
  "iconUrl": "https://example.com/icon.png",
  "supportedInterfaces": [
    {
      "url": "http://127.0.0.1:9001/invoke",
      "protocolBinding": "JSONRPC",
      "protocolVersion": "1.0"
    },
    {
      "url": "http://127.0.0.1:9001/",
      "protocolBinding": "HTTP+JSON",
      "protocolVersion": "1.0"
    }
  ],
  "defaultInputModes": ["text"],
  "defaultOutputModes": ["text", "file"],
  "capabilities": {
    "streaming": true,
    "pushNotifications": false,
    "extendedAgentCard": false
  },
  "securitySchemes": {
    "bearerAuth": {
      "type": "httpAuthSecurityScheme",
      "scheme": "bearer",
      "bearerFormat": "JWT"
    }
  },
  "skills": [
    {
      "id": "summarize_document",
      "name": "Summarize Document",
      "description": "Generates concise or detailed summaries of text documents or attached files.",
      "tags": ["summary", "nlp", "documents"],
      "examples": [
        "Summarize the key points from this report.",
        "Give me a 3-bullet summary of the attached document."
      ],
      "inputModes": ["text", "file"],
      "outputModes": ["text"]
    }
  ]
}
```

---

## 2. Key Authoring Rules for Discoverability

1. **Version Format:** Use `"1.0"`. Do NOT include patch numbers (e.g. `"1.0.0"` is forbidden by §3.6).
2. **Supported Interfaces Order:** Put your preferred transport first. `a2acli` and orchestrators select the first supported transport in `supportedInterfaces`.
3. **Intent-Rich Descriptions:** Write `description` strings aimed at LLM tool callers. Explain *what* the skill does and *when* to choose it over other skills.
4. **Concrete Examples:** Always supply 2–5 natural language user prompts in `skills[].examples`.
5. **Exact Match Capabilities:** Only declare `capabilities.streaming = true` or `capabilities.pushNotifications = true` if the underlying handlers actually support them.
