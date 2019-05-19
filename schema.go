package main

const schemaJSON = `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "title": "post",
  "description": "a blog post",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 50,
      "pattern": "^[A-Z].*"
    },
    "date": {
      "type": "string"
    },
    "body": {
      "type": "string"
    },
    "views": {
      "type": "integer",
      "minimum": 1
    },
    "post_type": {
      "type": "string",
      "enum": ["cross-post", "original"]
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  },
  "required": ["title", "date", "body", "author_email", "post_type"]
}`
