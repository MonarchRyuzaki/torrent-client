# Go BitTorrent Client 🚀

> "What I cannot create, I do not understand." — Richard Feynman

This is a custom BitTorrent client built from scratch in Go. It supports the core BitTorrent v1.0 protocol, including handshakes, bitfield exchange, pipelining, and piece verification.

The goal of this project was not just to download a file, but to engineer a robust Concurrent Distributed System that can handle the chaos of real-world networking (timeouts, choked peers, and dropped connections).


## 📖 The Story Behind This Project

Ever since I was young, I've been fascinated by the mechanics of data transfer. I remember the frustration of traditional browser downloads: no matter the file size, if the browser closed or the computer turned off, the progress was lost. You were effectively tethered to that active session.

BitTorrent was a revelation to me. It offered something unique: **resilience**. The ability to pause, resume, and pick up exactly where you left off felt like a superpower compared to the fragility of standard HTTP downloads. Understanding how this worked and building a tool that could handle this logic has always been on my bucket list.

As I began diving deeper into Go, I decided to finally tackle this challenge. I built this implementation over a weekend, and it served as a fantastic crash course in parsing binary formats and handling network peers. It is my first project of 2026, and I'm incredibly satisfied with how it turned out.

---

## 🏗️ The Architecture: "The Manager" Pattern

I chose a Central Dispatcher (Push) architecture over a Worker Queue (Pull).

### The Core Components

**The Download Manager (DownloadManager)**:

- Acts as the "Brain." It holds the global DownloadState (which pieces are done, busy, or pending).
- It loops constantly, identifying available pieces and assigning them to specific, free peers.
- **Bitfield-Aware Assignment**: Before assigning a piece, the manager checks each peer's bitfield (`peer.Bitfield.HasPiece(index)`) to ensure the peer actually has that piece. This prevents wasted work.
- Why this choice? It gives me granular control to implement strategies like "Rarest First" or "Endgame Mode" in the future.

**The Workers (Goroutines)**:

- Every download job is spawned as a lightweight Goroutine.
- Fire-and-Forget: The Manager launches a worker and immediately continues to the next piece. It does not wait for the download to finish.

**State Management**:

- **Global Mutex** (`ds.mu`): Protects the shared DownloadState (piece status slice, counter for remaining pieces).
- **Per-Peer Mutexes** (`pc.mu[peer_index]`): Each peer connection has its own lock, allowing multiple peers to download different pieces in parallel without blocking each other. This maximizes concurrency—only peers competing for the same connection are serialized.

### The Flow

```
[Manager Loop] 
   |---> Finds Pending Piece #5
   |---> Scans Peer List for Free Peer
   |---> Checks Peer A's Bitfield: HasPiece(5)? ✓
   |---> Locks Peer A's Mutex
   |---> Marks Peer A as Busy & Piece #5 as In-Progress
   |---> Spawns Goroutine [Peer A downloads Piece #5]
   |---> Unlocks & Loops immediately to Piece #6
   
   [Worker Goroutine (Peer A, Piece #5)]
   |---> Downloads Piece #5
   |---> On Success: Locks ds.mu → Marks Piece #5 Done → Frees Peer A
   |---> On Failure: Locks ds.mu → Resets Piece #5 → Kills Peer A
```

## ⚔️ Bugs & Trade-offs

The hardest part wasn't the protocol spec—it was the implementation details. Here are the specific concurrency battles I fought.

### 🐛 Bug 1: Sequential Concurrency

**Symptom**: I used Goroutines, but logs showed Piece #1 waiting for Piece #0 to finish.

**The Issue**: I was calling the blocking `DownloadPiece()` function inside the Manager loop. The loop itself was paused.

**The Fix**: Wrapped the download logic in a `go func(...)` closure. This unblocked the Manager, allowing it to assign 50 pieces in milliseconds.

### 🐛 Bug 2: The "Infinite Piece 0" Loop

**Symptom**: Logs showed "Piece 0 Done" repeatedly, but the download never progressed past the first piece.

**The Issue**: After assigning Piece 0 to a peer, the Manager loop would immediately continue to the next iteration... which started again at `piece_index = 0`. Without breaking out of the inner peer loop, it would reassign Piece 0 to another peer, and another, until all peers were downloading the same piece.

**The Fix**: Introduced an `assigned` flag. When a piece is successfully assigned, set `assigned = true` and `break` out of the peer loop. Then check `if assigned { continue }` to skip to the next piece instead of re-checking the current one.

### 🐛 Bug 3: The "Zombie Peer" Loop

**Symptom**: Infinite `write: broken pipe` logs.

**The Issue**: When a peer disconnected, my error handler caught it but marked the peer as Ready (0). The Manager saw a "Ready" peer and immediately tried to reuse the dead connection.

**The Fix**: Implemented ruthless state management. If a peer errors, I explicitly `Close()` the socket and mark the peer as Dead (2), permanently banning it from the pool.

### 🐛 Bug 4: The "Time Bomb" Timeout

**Symptom**: Downloads started fast but failed consistently after 5 seconds with `i/o timeout`.

**The Issue**: I set a 5-second deadline for the initial Handshake (`conn.SetDeadline`) but never turned it off. The OS faithfully killed the connection 5 seconds later.

**The Fix**: Added `conn.SetDeadline(time.Time{})` immediately after a successful handshake to disable the timer.


### 🐛 Bug 5: The "Nil Pointer Panic"

**Symptom**: Runtime panic: `invalid memory address or nil pointer dereference` when closing connections.

**The Issue**: When a peer handshake failed, the connection object (`conn`) was `nil`, but my cleanup code blindly called `conn.Close()` on every peer. Go doesn't gracefully handle `nil.Close()`—it panics.

**The Fix**: Added a nil check before closing: `if pc.pConn[peer_index] != nil { pc.pConn[peer_index].Close() }`. A simple defensive programming practice that saved me from random crashes during torrent downloads with flaky peers.


## ⚖️ Trade-offs & Limitations

To ship a robust v1.0, I made specific engineering trade-offs:

- **TCP Only**: I ignored UDP (uTP/DHT) to focus on reliable data transfer.
- **Single-File Support**: Multi-file torrents (folders) are currently ignored to simplify the Bencode parser.
- **Read Deadlines**: Added strict 30s read deadlines to prevent "Silent Peers" from starving a worker thread forever.
- **Static Peer Pool**: I only use the initial peer list from the tracker's first response. The `interval` field (which tells when to re-announce) is ignored, meaning I don't discover new peers or refresh the swarm during a download.
- **Batch Handshake Initialization**: All handshakes happen upfront in `GetPeerConnection()` using a `sync.WaitGroup`. The download only starts after *all* handshakes complete (or fail). This creates an artificial delay—a better approach would be to start downloading as soon as the first peer becomes ready, and continue handshaking with others in parallel.

These choices allowed me to focus on the **core distributed systems challenges**: concurrent downloads, state management, and error handling. For a learning project built over a weekend, getting the fundamentals right was more important than implementing every optimization. These features can be added incrementally in future iterations.

---
## 📚 Resources & Acknowledgments

This implementation follows the [BitTorrent Protocol Specification (BEP 3)](https://www.bittorrent.org/beps/bep_0003.html).

Special thanks to:
- **[CodeCrafters](https://app.codecrafters.io/courses/bittorrent/overview)** for providing the structured challenge that kickstarted this project
- **[Jesse Li's "Building a BitTorrent client from the ground up in Go"](https://blog.jse.li/posts/torrent/)** for the excellent technical breakdown that clarified many protocol subtleties

Built with ❤️ (and a lot of printf) by Shivam Ganguly.