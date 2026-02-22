# Create a Publish Subscribe HTTP Server

httpQ is a simple broker that provides a number of HTTP "channels" (topics).

## Behavior

### Producer (POST /{topic})
- Publishes a message to a topic
- If no consumers are waiting, **blocks** until a consumer connects on that channel
- Times out after **30 seconds** if no consumer arrives

### Consumer (GET /{topic})
- Returns a message published to a topic
- If no producers are waiting, **blocks** until a producer connects on that channel
- Times out after **30 seconds** if no producer arrives

### Ordering
- Messages should be delivered in **FIFO order** per topic
- No message persistence is required

## Stats Endpoint (GET /stats)

Return a JSON object with the following fields:

| Field | Description |
|-------|-------------|
| `rx_bytes` | Total bytes consumed (message bodies received by consumers) |
| `tx_bytes` | Total bytes published (message bodies sent by producers) |
| `pub_fails` | Number of publish failures (e.g., timeout) |
| `sub_fails` | Number of subscribe/consume failures (e.g., timeout) |

## Examples

**Producer:**
```bash
curl -k https://localhost:24744/NhPvrxcJ5WfsYJ -d "hello 1"
curl -k https://localhost:24744/NhPvrxcJ5WfsYJ -d "hello 2"
```

**Consumer:**
```bash
curl -k https://localhost:24744/NhPvrxcJ5WfsYJ
# Output: hello 1

curl -k https://localhost:24744/NhPvrxcJ5WfsYJ
# Output: hello 2
```

**Stats:**
```bash
curl -k https://localhost:24744/stats
# Output: {"rx_bytes":14,"tx_bytes":14,"pub_fails":0,"sub_fails":0}
```

## Deliverables

1. Create a new project with git
2. Setup Go module
3. Create binary with HTTPS server and self-signed certificates
4. Complete tests to verify functionality

## Evaluation Criteria

- [ ] **Correctness**: Pub/Sub blocking behavior works as specified
- [ ] **Concurrency safety**: No race conditions under concurrent load
- [ ] **Code organization**: Clean, readable, idiomatic Go
- [ ] **Error handling**: Timeouts and edge cases handled gracefully
- [ ] **Test coverage**: Meaningful tests that verify behavior
- [ ] **HTTPS**: Self-signed certificates correctly configured
- [ ] **Stats accuracy**: Counters are thread-safe and accurate

## Notes

- Time of commits is not considered; we acknowledge that you might not work on it all at once
- Time to complete is not considered
- We expect this to take approximately **2–4 hours** for a complete solution