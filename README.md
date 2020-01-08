# Pest-control
📬 Service for handling notification preferences

## API Documentation
The following APIs are protected by `heimdall`, so requests must have the
`Authorization` header set to the value `Bearer <token>`, where `<token>` is the
token generated by `heimdall`. `heimdall` would then forward the request with an
added `user_id` value that would be taken from the token.

### POST api/prefs
Creates a new set of preferences for a user.

#### Request body format
```
{
    "global": {
        "invitation": boolean (default: true),
        "text_entered": boolean (default: true),
        "text_modified": boolean (default: true),
        "tag": boolean (default: true),
        "role": boolean (default: true),
    },
    "conversation": [
        {
            "conversation_id": string (default: ""),
            "text_entered": boolean (default: true),
            "text_modified": boolean (default: true),
            "tag": boolean (default: true),
            "role": boolean (default: true),
        }
    ]
}
```

By default (i.e. if the request body is `{}`), all `global` fields are `true`
and there are no conversation notification preferences.
