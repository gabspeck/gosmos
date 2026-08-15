# MOS Remote Procedure Call Protocol Specification

## 1 Introduction

The MOS Remote Procedure Call Protocol (MOS RPC) carries remote procedure calls
between a client application and an online service over a single point-to-point
byte stream. It provides its own framing, error detection, retransmission, and
multiplexing, and does not depend on any facility of the underlying link beyond
ordered delivery of bytes.

The protocol is organized in three layers, all defined in this document:

- **Packet layer.** Frames the byte stream into packets, detects corruption, and
  recovers lost packets with a sliding window.
- **Pipe layer.** Multiplexes up to 16 independent logical channels (pipes) onto
  one connection, and fragments messages that exceed the packet size.
- **Call layer.** Carries requests and replies as host blocks addressed to an
  interface and a method, with typed parameters.

### 1.1 Glossary

**call** — a request sent by a client on a pipe, together with the reply the
server returns for it.

**call block** — a host block that carries a call or a reply.

**check field** — the four-byte error-detection field at the end of a packet.

**chunked field** — a variable-length parameter too large to travel inside a
call block, sent instead as a separate sequence of stream frames.

**client** — the peer that establishes the connection, opens pipes, and issues
calls.

**connection** — one instance of the protocol running over one byte stream.

**dynamic section** — the trailing part of a reply body whose content is opaque
to the call layer and runs to the end of the host block.

**host block** — the unit of the call layer: a class byte, a method byte, a
request identifier, and a body.

**interface** — a named set of methods, identified by a GUID and addressed on
the wire by a one-byte interface identifier assigned by the server.

**interface table** — the message that maps interface GUIDs to interface
identifiers on a newly opened pipe.

**method** — one operation of an interface, addressed by a one-byte method
identifier.

**packet** — the unit of the packet layer, terminated by the byte 0x0D.

**peer** — either party to a connection.

**pipe** — one of 16 logical channels multiplexed on a connection. Pipe 0 is
reserved for connection control; pipes 1 through 15 each carry one service.

**pipe frame** — the unit of the pipe layer: a header byte, an optional length
byte, and content.

**pipe message** — a complete unit of pipe content, possibly assembled from
several pipe frames carried in several packets.

**receive descriptor** — a tag in a request body that declares the type of one
field the caller expects in the reply.

**reply** — a host block returned for a call, repeating the call's class,
method, and request identifier.

**request identifier** — the value that distinguishes concurrent calls on one
pipe.

**routing value** — the two-byte value at the start of a pipe message that
selects its destination.

**send parameter** — a tag and value in a request body that carries an input
argument.

**server** — the peer that answers pipe-open requests, publishes interface
tables, and executes calls.

**service** — a named endpoint a client opens a pipe to, exposing one or more
interfaces.

**static section** — the leading part of a reply body, made of typed fields that
correspond positionally to the call's receive descriptors.

**stream frame** — a one-way host block carrying part of a chunked field.

### 1.2 References

#### 1.2.1 Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement
Levels", BCP 14, RFC 2119, March 1997.

[RFC4122] Leach, P., Mealling, M., and Salz, R., "A Universally Unique
IDentifier (UUID) URN Namespace", RFC 4122, July 2005. Referenced only for the
textual form of a GUID; MOS RPC transmits GUIDs in the little-endian structure
layout defined in section 2.2.1.4.

#### 1.2.2 Informative References

None.

### 1.3 Overview

A connection begins when the client establishes a byte stream to the server and
completes link initialization. The server then sends its transport parameters,
the client confirms the connection, and both peers exchange packets under a
sliding window.

All further traffic is multiplexed as pipe messages. Pipe 0 carries connection
control: transport parameters, connection request and confirmation, pipe-open
requests, and their responses. To reach a service, the client sends a pipe-open
request on pipe 0 naming the service, a version, and the pipe index it intends
to use. The server answers on that index and immediately sends the interface
table for the service, mapping each interface GUID the service supports to a
one-byte interface identifier.

The client resolves the GUID it needs against that table and issues calls. A
call is a host block whose class byte is the interface identifier, whose method
byte selects the operation, and whose body holds send parameters followed by
receive descriptors. The server executes the method and returns a reply that
repeats the class, method, and request identifier, and whose body holds the
fields the receive descriptors asked for.

A reply body has two parts. The static section carries fixed-size and
variable-length fields and ends with an end-of-static tag. The dynamic section,
if present, carries an opaque byte range that runs to the end of the host block;
its terminating tag tells the client whether the request is now complete or
whether more messages follow on the same request identifier. That mechanism
carries result sets, file transfers, and any payload larger than one reply.

Arguments too large to fit in a call block travel the other way, as chunked
fields: the call block carries a six-byte reference and the bytes follow in
one-way stream frames on the same pipe.

```
  application                calls, replies, notifications
  ---------------------------------------------------------
  call layer        host blocks, interfaces, methods, parameters
  ---------------------------------------------------------
  pipe layer        16 logical channels, fragmentation, reassembly
  ---------------------------------------------------------
  packet layer      framing, escape encoding, check field, window
  ---------------------------------------------------------
  byte stream       serial link or TCP connection
```

### 1.4 Relationship to Other Protocols

MOS RPC requires an ordered byte stream. It runs over a serial link or over TCP,
and takes nothing from either beyond the stream itself: framing, integrity, and
retransmission are provided by the packet layer whether or not the lower layer
also provides them.

Services carried over MOS RPC define their own message contracts on top of the
call layer. Those contracts are out of scope for this document, which specifies
the framing, addressing, and parameter encoding every one of them uses.

### 1.5 Prerequisites/Preconditions

A byte stream to the server endpoint must already exist. Endpoint selection and
link establishment are outside this specification.

Both peers must be able to transmit and receive all 256 byte values except that
the packet layer reserves 0x0D as a terminator and escapes it inside packets;
implementations therefore interoperate over links that are transparent to eight
bits.

### 1.6 Applicability Statement

MOS RPC suits request-response and streaming interaction between one client and
one server over a slow, error-prone link. Its window, packet size, and per-pipe
reassembly are sized for links of a few kilobytes per second. It provides no
confidentiality or integrity guarantee against an active attacker; see section
5.

### 1.7 Versioning and Capability Negotiation

The protocol negotiates in three places:

- **Transport parameters.** The server sends its packet size, window, and timer
  values (section 2.2.4.1.2). The client applies the smaller of each server
  value and its own configured maximum.
- **Service version.** The pipe-open request names a service version (section
  2.2.4.2). The server accepts or refuses the pipe.
- **Interfaces.** The interface table (section 2.2.7) tells the client which
  interfaces the service exposes on this connection and what identifier each has
  on this pipe. Identifiers are per-pipe assignments, not fixed constants; a
  client MUST resolve them from the table on every pipe it opens.

There is no protocol version number on the wire.

### 1.8 Vendor-Extensible Fields

Interface identifiers 0x01 through 0xDF are assigned by the server per pipe and
carry no meaning outside a connection. Services extend the protocol by defining
new interface GUIDs and publishing them in the interface table.

The class range 0xE0 through 0xFF is reserved by this specification (section
2.2.6). The tag values in sections 2.2.8 and 2.2.9 are reserved; a receiver
that meets an unassigned tag cannot determine the length of the field that
follows it and MUST stop parsing the body at that point.

### 1.9 Standards Assignments

None. Interface identity uses GUIDs, which require no assignment authority.

## 2 Messages

### 2.1 Transport

MOS RPC uses one bidirectional byte stream per connection. Both peers may
transmit at any time after link initialization completes.

The stream carries packets (section 2.2.2). Every packet ends with the byte
0x0D, which never occurs anywhere else in a packet. A receiver therefore frames
the stream by scanning for 0x0D, and MUST treat everything since the previous
terminator as one packet.

### 2.2 Message Syntax

#### 2.2.1 Common Data Types

##### 2.2.1.1 Byte Order and Bit Numbering

Multi-byte integers are little-endian unless a field's definition states
otherwise. Two fields are big-endian and are marked as such: the variable-length
integer (section 2.2.1.2) and the extended form of the variable-length field
size (section 2.2.1.3).

Bit 0 is the least significant bit of a byte. Byte offsets in layout diagrams
start at 0.

##### 2.2.1.2 Variable-Length Integer (VLI)

A VLI encodes an unsigned integer in one, two, or four bytes. The two most
significant bits of the first byte select the form; the remaining bits are the
value, most significant bits first.

| Bits 7-6 of byte 0 | Size | Value | Range |
|---|---|---|---|
| 00 | 1 byte | `b0 & 0x3F` | 0 - 63 |
| 10 | 2 bytes | `((b0 & 0x3F) << 8) \| b1` | 0 - 16,383 |
| 11 | 4 bytes | `(b0..b3 as big-endian uint32) & 0x3FFFFFFF` | 0 - 1,073,741,823 |

A sender MUST use the shortest form that holds the value. A receiver MUST accept
the encoding `01` in bits 7-6 as equivalent to `00`.

Examples:

| Value | Encoding |
|---|---|
| 0 | `00` |
| 63 | `3f` |
| 64 | `80 40` |
| 16,383 | `bf ff` |
| 16,384 | `c0 00 40 00` |
| 1,073,741,823 | `ff ff ff ff` |

##### 2.2.1.3 Variable-Length Field Size

Variable-length parameter fields (sections 2.2.8.1 and 2.2.9.3) prefix their
content with a size in one of two forms, selected by bit 7 of the first byte.

| Bit 7 of byte 0 | Size | Value | Range |
|---|---|---|---|
| 1 | 1 byte | `b0 & 0x7F` | 0 - 127 |
| 0 | 2 bytes | `(b0 << 8) \| b1`, big-endian | 0 - 32,767 |

A sender MUST use the one-byte form for lengths below 128. A field longer than
32,767 bytes MUST be sent as a chunked field (section 2.2.8.2).

##### 2.2.1.4 Interface Identifier GUID

A GUID travels as 16 bytes in structure layout: a 32-bit field, two 16-bit
fields, and eight bytes, where the first three fields are little-endian and the
final eight bytes are in order.

The GUID `00028BB6-0000-0000-C000-000000000046` encodes as:

```
b6 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46
```

Receivers compare GUIDs byte for byte.

#### 2.2.2 Packet

```
 0        1        2                          n-5      n-1  n
+--------+--------+--------------------------+--------+----+
| Seq    | Ack    | Payload (escape-encoded) | Check  |0x0D|
+--------+--------+--------------------------+--------+----+
    1        1              0 .. N                4      1
```

| Field | Size | Description |
|---|---|---|
| Seq | 1 | Packet type and sequence number. See below. |
| Ack | 1 | Acknowledgment. Bit 7 set; bits 6-0 are the sequence number the sender next expects from its peer. |
| Payload | 0 - N | Pipe frames (section 2.2.3), escape-encoded per section 2.2.2.1. |
| Check | 4 | Check field over bytes 0 through n-6, masked per section 2.2.2.3. |
| Terminator | 1 | 0x0D. |

The Seq byte selects the packet type:

| Value | Type |
|---|---|
| 0x41 | Acknowledgment packet (section 2.2.2.4) |
| 0x42 | Negative acknowledgment packet (section 2.2.2.5) |
| any other | Data packet |

In a data packet, Seq has bit 7 set and bits 6-0 hold the sequence number, after
the reserved-value transform of section 2.2.2.2 is reversed. Sequence numbers
run 0 through 127 and wrap.

The minimum packet is seven bytes: Seq, Ack, an empty payload, the check field,
and the terminator. The maximum is the negotiated packet size (section
2.2.4.1.2), which counts every byte of the packet including the terminator.

An empty packet — a terminator with nothing before it — is used during link
initialization (section 3.1.3) and carries no other meaning.

##### 2.2.2.1 Escape Encoding

The payload is escape-encoded so that no byte of it collides with the terminator
or with the byte values the link reserves. The escape byte is 0x1B. Each of the
following raw values is replaced by a two-byte sequence:

| Raw | Encoded | Raw | Encoded |
|---|---|---|---|
| 0x1B | `1b 30` | 0x0B | `1b 33` |
| 0x0D | `1b 31` | 0x8D | `1b 34` |
| 0x10 | `1b 32` | 0x90 | `1b 35` |
| | | 0x8B | `1b 36` |

All other values are transmitted unchanged. Decoding replaces each escape byte
and its successor with the corresponding raw value; an escape byte followed by
any other value decodes to that value.

Encoding applies to the payload only. The Seq and Ack bytes use the transform in
section 2.2.2.2, and the check field uses the mask in section 2.2.2.3.

The encoding expands the payload, at worst to twice its length. A sender that
fragments a message MUST size each fragment against the encoded length, because
expansion depends on content (section 3.1.5.3).

##### 2.2.2.2 Reserved-Value Transform

The Seq and Ack bytes, and the header and length bytes of a pipe frame (section
2.2.3), are not escape-encoded. They avoid the reserved values by a reversible
transform instead: if the byte would be 0x8D, 0x90, or 0x8B, it is transmitted
exclusive-ORed with 0xC0, producing 0x4D, 0x50, or 0x4B.

A receiver MUST reverse the transform: a received 0x4D, 0x50, or 0x4B in one of
these positions is exclusive-ORed with 0xC0 to recover the value. All other
values pass through unchanged in both directions.

The transform applies to the byte after bit 7 has been set, so a sender computes
the Seq byte as `transform(sequence | 0x80)` and the Ack byte as
`transform(acknowledgment | 0x80)`.

##### 2.2.2.3 Check Field

The check field is a 32-bit cyclic redundancy check computed over the packet
bytes as they appear on the wire — the Seq byte, the Ack byte, and the encoded
payload — and transmitted little-endian, then masked.

The register is initialized to 0, no final inversion is applied, and the
polynomial is 0x248EF9BE with an inversion on the odd path:

```
for each of the 8 bits of each input byte, low bit first:
    if (register & 1) != 0:
        register = ~((register ^ 0x248EF9BE) >> 1)
    else:
        register = register >> 1
```

which yields the table-driven form:

```
table[i] = the value above, seeded with register = i
crc      = 0
for each byte b of the input:
    crc = (crc >> 8) ^ table[(b ^ (crc & 0xFF)) & 0xFF]
```

All arithmetic is on 32-bit unsigned values.

The four bytes of the little-endian result are then masked: any byte whose value
is 0x1B, 0x0D, 0x10, 0x0B, 0x8D, 0x90, or 0x8B is transmitted OR 0x60. The mask
is not reversible. A receiver MUST compute the check value over the received
bytes, apply the same mask to its own result, and compare the masked forms.

Worked values:

| Packet bytes checked | Masked check field |
|---|---|
| `80 80 e0 03 00 ff ff 04` | `fc 03 18 a0` |
| `80 80 e0 17 00 ff ff 03 00 04 00 00 00 04 00 00 1b 32 00 00 00 01 00 00 00 58 02 00 00` | `38 c9 9a 7e` |

##### 2.2.2.4 Acknowledgment Packet

```
+--------+--------+--------+------+
|  0x41  |  Ack   | Check  | 0x0D |
+--------+--------+--------+------+
```

An acknowledgment packet carries no payload and no sequence number. Its Ack
field has bit 7 set and bits 6-0 hold the next sequence number the sender
expects. It is not itself acknowledged and does not advance the sender's
sequence space.

##### 2.2.2.5 Negative Acknowledgment Packet

```
+--------+--------+--------+------+
|  0x42  | Seq    | Check  | 0x0D |
+--------+--------+--------+------+
```

Byte 1 masked with 0x7F is the last sequence number the sender received intact.
The peer resumes transmission from the following sequence number.

#### 2.2.3 Pipe Frame

A packet payload holds one or more pipe frames laid end to end.

```
 0            1              1 or 2
+------------+--------------+------------------+
| PipeHeader | Length (opt) | Content          |
+------------+--------------+------------------+
```

PipeHeader, after the reserved-value transform of section 2.2.2.2 is reversed:

| Bits | Mask | Field | Description |
|---|---|---|---|
| 3-0 | 0x0F | PipeIndex | 0 - 15. |
| 4 | 0x10 | HasLength | A length byte follows the header. |
| 5 | 0x20 | Continuation | The frame runs to the end of the packet or to the end of its content, whichever comes first. |
| 6 | 0x40 | LastData | This frame ends the pipe message. |
| 7 | 0x80 | — | Always set. |

Exactly one of HasLength and Continuation MUST be set.

When HasLength is set, the byte after the header — also subject to the transform
of section 2.2.2.2 — has bit 7 set, and bits 6-0 give the number of content
bytes that follow, at most 127.

Content of the first frame of a pipe message:

```
+--------------+---------------------------+
| ContentLength|  Message bytes            |
+--------------+---------------------------+
       2                 0 .. 65535
```

ContentLength is a little-endian 16-bit count of the message bytes of the
**whole** pipe message, across every frame that carries it. Frames after the
first carry message bytes only, with no length field of their own.

A pipe message is complete when ContentLength bytes have arrived, and a receiver
MUST use that count — not the LastData flag — to decide where one message ends
and the next begins. LastData marks the final frame of a transmission sequence
and can be clear on the frame that completes a message.

#### 2.2.4 Pipe 0 Content

Every pipe message begins with a two-byte little-endian routing value that
selects its kind and destination.

| Routing value | Content |
|---|---|
| 0x0000 | Pipe-open request (section 2.2.4.2) |
| 0xFFFF | Control frame (section 2.2.4.1) |
| 0x0001 - 0x000F | Data for that pipe (section 2.2.4.4) |

##### 2.2.4.1 Control Frame

```
+--------------+--------+-----------------+
| 0xFFFF       | Type   | Type-specific   |
+--------------+--------+-----------------+
       2           1          0 .. N
```

| Type | Name | Direction |
|---|---|---|
| 1 | Connection request | client to server, echoed back |
| 3 | Transport parameters | server to client |
| 4 | Connection established | client to server |

###### 2.2.4.1.1 Connection Request (Type 1)

The type-specific field is opaque to the server, which MUST return a type 1
control frame whose type-specific field repeats the received bytes exactly.

###### 2.2.4.1.2 Transport Parameters (Type 3)

```
+-----------+-----------+-----------+-----------+-----------+-------------+
| PacketSize| MaxBytes  | WindowSize| AckBehind | AckTimeout| KeepAlive   |
+-----------+-----------+-----------+-----------+-----------+-------------+
      4           4           4           4           4        0 or 4
```

All fields are little-endian 32-bit unsigned integers.

| Field | Description |
|---|---|
| PacketSize | Maximum packet length in bytes, counting Seq, Ack, encoded payload, check field, and terminator. |
| MaxBytes | Maximum bytes the sender will hold outstanding. |
| WindowSize | Maximum unacknowledged packets in flight. |
| AckBehind | Number of unacknowledged received packets after which the receiver MUST send an acknowledgment packet. |
| AckTimeout | Retransmission timer, in milliseconds. |
| KeepAlive | Idle interval in milliseconds. Present only when the control frame body exceeds 0x18 bytes. |

The receiving peer applies the smaller of each received value and its own
configured maximum.

###### 2.2.4.1.3 Connection Established (Type 4)

The type-specific field is empty. The client sends this frame once it has
applied the transport parameters.

##### 2.2.4.2 Pipe Open Request

```
+--------+--------+-----------+-------------+-----------+---------+
| 0x0000 |Reserved| PipeIndex | ServiceName | Parameter | Version |
+--------+--------+-----------+-------------+-----------+---------+
    2        2          2        variable      variable      4
```

| Field | Size | Description |
|---|---|---|
| Routing | 2 | 0x0000. |
| Reserved | 2 | Set to 0 by senders; ignored by receivers. |
| PipeIndex | 2 | The pipe the client will use for this service, 1 - 15. |
| ServiceName | variable | Service name, ASCII, NUL-terminated. Matched case-insensitively. |
| Parameter | variable | Service-defined parameter string, ASCII, NUL-terminated. May be empty. |
| Version | 4 | Service version, little-endian. |

##### 2.2.4.3 Pipe Open Response

The response is sent as a pipe message whose routing value is the pipe index
being opened.

```
+-----------+---------+----------------+--------+
| PipeIndex | Command | ServerPipeIndex| Status |
+-----------+---------+----------------+--------+
      2          2           2              2
```

| Field | Size | Description |
|---|---|---|
| PipeIndex | 2 | Routing value: the pipe index from the request. |
| Command | 2 | 0x0001, pipe opened. |
| ServerPipeIndex | 2 | The index the server will use. Equal to PipeIndex. |
| Status | 2 | 0x0000 on success. |

The response MUST be sent in a frame with Continuation set. Sending it with
HasLength set instead causes the client to misroute it.

##### 2.2.4.4 Pipe Data

```
+-----------+-------------------------+
| PipeIndex | Host block              |
+-----------+-------------------------+
      2             variable
```

The routing value is the destination pipe index, and the rest of the pipe
message is one host block (section 2.2.6).

#### 2.2.5 Pipe Close

A pipe message whose content after the routing value is the single byte 0x01
closes that pipe. No response is sent. Either peer may close a pipe it opened or
serves; after a close, both peers discard the pipe's reassembly state and any
calls outstanding on it.

#### 2.2.6 Host Block

The first byte of a host block is the Class, which selects the layout of the
rest.

| Class | Layout |
|---|---|
| 0x00 | Interface table (section 2.2.7) |
| 0x01 - 0xDF | Call block (section 2.2.6.1) |
| 0xE0 - 0xFF | Stream frame (section 2.2.6.2) |

##### 2.2.6.1 Call Block

```
+--------+--------+------------------+------------------+
| Class  | Method | RequestId (VLI)  | Body             |
+--------+--------+------------------+------------------+
    1        1          1, 2, or 4        variable
```

| Field | Size | Description |
|---|---|---|
| Class | 1 | Interface identifier, as assigned by the interface table on this pipe. |
| Method | 1 | Method identifier within the interface. |
| RequestId | 1 - 4 | VLI (section 2.2.1.2). Distinguishes concurrent calls on one pipe. |
| Body | variable | Request parameters (section 2.2.8) or reply parameters (section 2.2.9). |

A reply MUST repeat the Class, Method, and RequestId of the call it answers. A
client discards any reply that does not match an outstanding call in all three
(section 3.2.5.2).

##### 2.2.6.2 Stream Frame

```
+--------+----------+-------------------+
| Class  | StreamId | Field bytes       |
+--------+----------+-------------------+
    1         1            variable
```

| Class | Description |
|---|---|
| 0xE6 | Part of a chunked field; more frames follow for this stream. |
| 0xE7 | Final part of a chunked field for this stream. |

A stream frame carries no method and no request identifier: byte 1 is the stream
identifier and the field bytes begin at byte 2. It is one-way. A receiver MUST
NOT reply to it; a reply would carry a request identifier the peer has no
outstanding call for and would be discarded, and the sender does not wait for
one.

#### 2.2.7 Interface Table

```
+--------+--------+--------+---------------------------------+
|  0x00  |  0x00  |  0x00  | Record [1..n]                   |
+--------+--------+--------+---------------------------------+
```

Class, Method, and RequestId are all 0. The body is a sequence of 17-byte
records:

```
+----------------------------------+-----------+
| InterfaceGuid                    | Interface |
+----------------------------------+-----------+
                16                       1
```

| Field | Size | Description |
|---|---|---|
| InterfaceGuid | 16 | GUID in the layout of section 2.2.1.4. |
| Interface | 1 | The identifier to use as the Class byte of calls to this interface, 0x01 - 0xDF. |

The table is unordered. Identifiers are unique within a table and valid only for
the pipe the table arrived on.

#### 2.2.8 Request Parameters

A request body is a sequence of tagged fields: send parameters first, then
receive descriptors. A receiver MUST also accept the two interleaved.

Bit 7 of a tag distinguishes the two: clear for a send parameter, set for a
receive descriptor. Bits 3-0 give the type.

##### 2.2.8.1 Send Parameters

| Tag | Type | Field |
|---|---|---|
| 0x01 | byte | 1 byte |
| 0x02 | word | 2 bytes, little-endian |
| 0x03 | dword | 4 bytes, little-endian |
| 0x04 | variable | size (section 2.2.1.3) followed by that many bytes |
| 0x05 | chunked reference | 5 bytes, section 2.2.8.2 |
| 0x45 | chunked reference | 5 bytes, section 2.2.8.2. Equivalent on the wire to 0x05. |

At most 16 send parameters may appear in one request.

##### 2.2.8.2 Chunked Field Reference

A variable-length argument that does not fit in the call block is replaced in
the body by a reference, and its bytes follow in stream frames (section
2.2.6.2).

```
+--------+----------+---------------+
| Tag    | StreamId | Length        |
+--------+----------+---------------+
    1         1            4
```

| Field | Size | Description |
|---|---|---|
| Tag | 1 | 0x05 or 0x45. |
| StreamId | 1 | Identifies the stream frames that carry this field. Unique among the streams outstanding on a connection. |
| Length | 4 | Total field length in bytes, little-endian. |

Stream identifiers are allocated from a per-connection counter starting at 1, so
concurrent calls on one connection never share one.

A receiver MUST NOT parse the five bytes after a 0x05 or 0x45 tag as a
variable-length field. The StreamId byte would be read as a size and the rest of
the request consumed as content.

##### 2.2.8.3 Receive Descriptors

A receive descriptor is a tag with bit 7 set and no data. It declares the type
of one field the caller expects in the reply, in order.

| Tag | Declares |
|---|---|
| 0x81 | byte |
| 0x82 | word |
| 0x83 | dword |
| 0x84 | variable |
| 0x85 | dynamic |

At most 16 receive descriptors may appear in one request. The reply's fields
correspond to them positionally (section 2.2.9).

#### 2.2.9 Reply Parameters

A reply body is a static section, an end-of-static tag, and an optional dynamic
section.

```
+---------------------------+--------+---------------------------+
| Static fields             |  0x87  | Dynamic section           |
+---------------------------+--------+---------------------------+
```

##### 2.2.9.1 Static Section

The static section holds one field per receive descriptor of the corresponding
type, in the order the descriptors appeared.

| Tag | Type | Field |
|---|---|---|
| 0x81 | byte | 1 byte |
| 0x82 | word | 2 bytes, little-endian |
| 0x83 | dword | 4 bytes, little-endian |
| 0x84 | variable | size (section 2.2.1.3) followed by that many bytes |
| 0x8F | error | 4 bytes, little-endian (section 2.2.9.5) |

The static section may be empty.

##### 2.2.9.2 End-of-Static

The tag 0x87 has no field and ends the static section. It MUST be present in
every reply, including replies whose static section is empty and replies that
carry no dynamic section.

##### 2.2.9.3 Variable Fields

A 0x84 field carries its own size and may appear before or after other static
fields, wherever its descriptor appeared. Its size uses the encoding of section
2.2.1.3.

##### 2.2.9.4 Dynamic Section

A dynamic tag is followed by no size. Its content is every remaining byte of the
host block, and the sender therefore MUST place it last. The tag determines what
the content does:

| Tag | Content | Effect |
|---|---|---|
| 0x85 | appended to the message being assembled | The message stays open. The call stays outstanding. |
| 0x88 | appended to the message being assembled | The message is complete and delivered to the caller. The call stays outstanding. |
| 0x86 | appended to the message being assembled | The call is complete. |

The three tags drive two different completion paths in the caller, and they are
not interchangeable:

- A caller that waits for a single result is released by 0x86 only. A reply that
  ends its dynamic section with 0x88 leaves that caller blocked until the pipe
  closes.
- A caller that reads a sequence of messages — a result set, a file, a
  notification feed — is fed by 0x88, which closes one message and makes it
  readable. A reply that ends with 0x86 delivers nothing to that caller, which
  sees an empty sequence.

A sequence therefore ends with two host blocks on the same request identifier:
the last content-bearing block ending in 0x88, then a block whose body is
`87 86` and which carries no content. Without the second block the caller never
learns that the sequence has ended.

One dynamic message MUST NOT exceed 16,384 bytes of content. A sender with more
to deliver MUST split it into several host blocks, each ending its own message
with 0x88. Consecutive 0x85 blocks accumulate into one message and overrun the
receiver's message buffer at that limit.

##### 2.2.9.5 Error Field

A 0x8F field replaces the reply the call would have returned. Its four
little-endian bytes are:

| Value | Meaning |
|---|---|
| 0xE0000001 | Parameter type does not match the request |
| 0xE0000002 | Message is not a valid host block |
| 0xE0000003 | Parameter could not be added, or the send failed |
| 0xE0000004 | Method is not registered on this interface |
| 0xE0000005 | Memory allocation failed |
| 0xE0000006 | Internal error |
| 0xE0000007 | Invalid parameter |
| 0xE0000008 | Send error |
| 0xE0000009 | Chunked field could not be received |
| 0xE000000A | The service refused the call |

#### 2.2.10 Iterator Cancel

A call block whose body is the single byte 0x0F cancels the sequence outstanding
on its Class, Method, and RequestId. It is sent by the caller when it abandons a
sequence before the sequence ends.

The peer MUST answer with a call block repeating that Class, Method, and
RequestId, whose body is `87 88`.

No other body encoding produces a single 0x0F byte: every send parameter tag
carries data after it, and every receive descriptor has bit 7 set.

## 3 Protocol Details

### 3.1 Common Details

Both peers implement the packet and pipe layers identically. This section
specifies that behavior once; sections 3.2 and 3.3 add the call-layer behavior
specific to each role.

#### 3.1.1 Abstract Data Model

Per connection:

- **SendSequence**: the sequence number for the next data packet. Initialized
  to 0 and incremented modulo 128 for every data packet sent.
- **ExpectedSequence**: the sequence number the peer is expected to send next,
  transmitted in the Ack field. Initialized to 0.
- **UnacknowledgedPackets**: data packets sent and not yet acknowledged, in
  order.
- **TransportParameters**: PacketSize, MaxBytes, WindowSize, AckBehind,
  AckTimeout, and KeepAlive, as negotiated in section 2.2.4.1.2.
- **PendingEscape**: the held payload of a packet that ended with an unpaired
  escape byte, or empty.
- **PipeTable**: for each pipe index in use, its state and its reassembly
  context.

Per pipe:

- **ReassemblyBuffer**: message bytes accumulated so far.
- **BytesOutstanding**: message bytes still expected for the message being
  assembled. Zero when no message is open.

#### 3.1.2 Timers

- **Retransmission timer.** Runs while UnacknowledgedPackets is non-empty, with
  a period of AckTimeout.
- **Acknowledgment timer.** Optional. A receiver that does not acknowledge every
  packet immediately MUST acknowledge within AckTimeout, and MUST acknowledge
  after AckBehind unacknowledged packets have accumulated.
- **Keep-alive timer.** Present only when KeepAlive was negotiated. On expiry
  the peer sends an acknowledgment packet.

#### 3.1.3 Initialization

Link initialization runs before any packet is exchanged.

1. The client sends an empty packet — the single byte 0x0D.
2. The server responds with the three ASCII characters `COM` followed by 0x0D.
   This response selects binary packet framing on the link.
3. The server sends the transport parameters control frame (section 2.2.4.1.2)
   as a data packet with sequence number 0.

The client MUST abort the connection if the response in step 2 does not arrive
within 20 seconds, and MUST abort if it is anything other than `COM` followed by
0x0D.

After step 3 both peers are in the connected state and the packet layer is
active in both directions.

#### 3.1.4 Higher-Layer Triggered Events

**Sending a pipe message.** The pipe layer accepts a message for a pipe,
fragments it per section 3.1.5.3, and hands packets to the packet layer.

**Closing a pipe.** The peer sends the pipe-close message of section 2.2.5 and
discards the pipe's reassembly state.

#### 3.1.5 Message Processing Events and Sequencing Rules

##### 3.1.5.1 Sending a Packet

To send a data packet:

1. Take the sequence number from SendSequence and increment SendSequence modulo
   128.
2. Build the Seq byte as `transform(sequence | 0x80)` and the Ack byte as
   `transform(ExpectedSequence | 0x80)`, per section 2.2.2.2.
3. Escape-encode the payload per section 2.2.2.1.
4. Compute the check field over the Seq byte, the Ack byte, and the encoded
   payload; mask it per section 2.2.2.3.
5. Append the terminator 0x0D.
6. Append the packet to UnacknowledgedPackets and start the retransmission timer
   if it is not running.

A sender MUST NOT have more than WindowSize packets in UnacknowledgedPackets. A
sender MUST NOT emit a packet longer than PacketSize.

Acknowledgment packets are built the same way but take no sequence number, are
not added to UnacknowledgedPackets, and do not start the retransmission timer.

##### 3.1.5.2 Receiving a Packet

1. Read bytes until 0x0D. The bytes before it are the packet.
2. If fewer than six bytes were read, discard the packet.
3. Compute the check field over every byte except the last four; mask it;
   compare with the last four bytes. On mismatch, discard the packet and leave
   PendingEscape untouched. No error is signaled to the peer.
4. Reverse the transform of section 2.2.2.2 on bytes 0 and 1 to recover the
   packet type, the sequence number, and the acknowledgment.
5. Release from UnacknowledgedPackets every packet whose sequence number is
   before the received acknowledgment, and stop the retransmission timer if none
   remain.
6. For an acknowledgment or negative acknowledgment packet, stop here.
7. Set ExpectedSequence to the received sequence number plus 1, modulo 128, and
   send an acknowledgment packet. Do this before processing the payload, so that
   any reply generated by the payload carries the updated value.
8. Decode the payload per section 3.1.5.4 and process the pipe frames it
   contains.

##### 3.1.5.3 Fragmenting a Pipe Message

A pipe message that does not fit in one packet is split across consecutive
packets:

```
Frame 1     Continuation set, LastData clear
            PipeHeader | ContentLength | bytes[0..a]

Frame 2..k-1  Continuation set, LastData clear
            PipeHeader | bytes[a..b]

Frame k     Continuation set, LastData set
            PipeHeader | bytes[..end]
```

Rules a sender MUST observe:

- ContentLength on frame 1 is the length of the **entire** message. A frame-1
  length that covers only frame 1's own bytes fills the receiver's buffer early;
  the receiver then consumes no bytes from the following frame and retries the
  same input indefinitely.
- Frames after the first carry no length field. Bytes written there as a length
  are delivered to the application as message content.
- Each fragment goes in its own packet, with its own sequence number, in order.
- Fragment sizes MUST be chosen so that the **encoded** packet fits PacketSize.
  A fixed allowance for expansion is not sufficient: a fragment dense in the
  values of section 2.2.2.1 doubles in length, and a packet over PacketSize is
  discarded by the receiver without notice. A sender picks the largest fragment
  whose fully built packet fits.

##### 3.1.5.4 Decoding a Payload and Reassembling Messages

Escape pairs may straddle a packet boundary. Decoding therefore keeps state
across packets:

1. If PendingEscape is non-empty, the current packet's payload begins with a
   repeated pipe frame header, and the byte after it completes the held escape
   pair. Decode PendingEscape followed by that byte as one unit, clear
   PendingEscape, remove that byte from the payload, and continue with the
   header and the remaining bytes.
2. If the payload now ends with an unpaired escape byte, store the whole payload
   in PendingEscape and stop; it will be decoded when the next packet arrives.
3. Otherwise decode the payload per section 2.2.2.1.

For each pipe frame in the decoded payload:

1. Reverse the transform of section 2.2.2.2 on the header byte, and on the
   length byte if HasLength is set, to recover PipeIndex and the flags.
2. Determine the frame's content bytes: the bytes named by the length byte if
   HasLength is set, otherwise the rest of the payload.
3. If BytesOutstanding for the pipe is zero, read ContentLength from the first
   two content bytes and take up to that many of the bytes that follow.
   Otherwise take up to BytesOutstanding bytes, with no length field present.
4. Append the taken bytes to ReassemblyBuffer and reduce BytesOutstanding by the
   number taken.
5. If BytesOutstanding is now zero, the message is complete: deliver
   ReassemblyBuffer and clear it.
6. Advance to the next frame at the first byte the current frame did not consume.

A receiver MUST NOT assume a continuation frame owns the rest of the packet: a
second frame can follow it in the same payload, and consuming to the end of the
payload discards it.

##### 3.1.5.5 Routing a Pipe Message

Read the two-byte routing value at the start of the message and dispatch per
section 2.2.4. A message shorter than two bytes is discarded. A message for a
pipe that is not open is discarded.

The pipe index in the frame header and the routing value are independent. A
receiver MUST route on the routing value.

#### 3.1.6 Timer Events

**Retransmission timer expiry.** The sender retransmits every packet in
UnacknowledgedPackets, in sequence order, and halves its effective window to a
minimum of one packet. After 12 consecutive expiries without progress the sender
terminates the connection.

**Keep-alive timer expiry.** The peer sends an acknowledgment packet.

#### 3.1.7 Other Local Events

**Loss of the byte stream.** All pipes are closed, all outstanding calls fail,
and the connection is discarded.

### 3.2 Client Details

#### 3.2.1 Abstract Data Model

In addition to section 3.1.1:

- **ServicePipes**: for each open pipe, the service name, the version, and the
  interface map.
- **InterfaceMap**: for each pipe, the GUID-to-identifier assignments received in
  the interface table.
- **OutstandingCalls**: for each pipe, the calls awaiting completion, keyed by
  request identifier, each holding the Class and Method sent and the caller
  waiting on it.
- **NextRequestId**: per pipe, the request identifier for the next call.
- **NextStreamId**: per connection, the stream identifier for the next chunked
  field. Initialized to 1.

#### 3.2.2 Timers

None beyond section 3.1.2. A client MAY apply its own call timeout; the protocol
defines none.

#### 3.2.3 Initialization

The client performs steps 1 and 2 of section 3.1.3, receives the transport
parameters, and applies the smaller of each received value and its own
configured maximum. It then sends, on pipe 0:

1. A type 4 control frame (section 2.2.4.1.3).
2. A type 1 control frame (section 2.2.4.1.1) whose body it will compare against
   the echo.

The connection is ready when the echo returns with the body unchanged.

#### 3.2.4 Higher-Layer Triggered Events

**Opening a service.** The client allocates an unused pipe index, sends a
pipe-open request (section 2.2.4.2), and waits for the pipe-open response and
then the interface table. It MUST NOT issue a call on the pipe before it has
resolved the GUID it needs to an identifier from that table.

**Issuing a call.** The client:

1. Allocates a request identifier that no outstanding call on the pipe is using.
2. Builds the body: send parameters in the order the method defines, then one
   receive descriptor per field it expects back.
3. Replaces any variable-length argument that does not fit in the call block
   with a chunked reference (section 2.2.8.2), allocating a stream identifier
   from NextStreamId.
4. Sends the call block on the pipe, then the stream frames for each chunked
   field, in order, 0xE7 on the last frame of each stream.
5. Records the call in OutstandingCalls with the Class and Method it sent.

The client MUST NOT wait for a stream to be consumed before sending its next
message; stream frames are one-way and unacknowledged.

**Cancelling a sequence.** The client sends the iterator-cancel body (section
2.2.10) on the Class, Method, and RequestId of the sequence, and treats the call
as complete when the `87 88` acknowledgment arrives.

**Closing a service.** The client sends the pipe-close message (section 2.2.5)
and fails any calls still outstanding on that pipe.

#### 3.2.5 Message Processing Events and Sequencing Rules

##### 3.2.5.1 Receiving an Interface Table

The client stores the GUID-to-identifier assignments for the pipe. An identifier
is meaningful only on that pipe.

If a GUID the client requires is absent from the table, no interface can be
resolved and the client abandons the pipe. A server that omits a GUID a service
needs therefore causes the client to close the pipe immediately after the table
arrives, without issuing a call.

##### 3.2.5.2 Receiving a Reply

A reply is matched in three steps:

1. Class against the interface handler registered for the pipe.
2. RequestId against OutstandingCalls.
3. Method against the Method recorded with that outstanding call.

A reply that fails any step is discarded. No error is returned and no indication
reaches the application. A server that alters any of the three values in its
reply produces a call that never completes.

On a match, the client parses the body per section 2.2.9:

- Static fields are delivered to the caller in order.
- The dynamic tag determines completion, per the table in section 2.2.9.4. On
  0x86 the call completes and is removed from OutstandingCalls. On 0x88 the
  current message is delivered and the call remains outstanding. On 0x85 the
  content is accumulated and the call remains outstanding.
- A 0x8F field fails the call with the value it carries.

##### 3.2.5.3 Receiving an Unsolicited Message

A server may send a call block on a pipe with the Class, Method, and RequestId
of a call the client left outstanding — that is how a sequence delivers messages
after the first. The client processes it exactly as section 3.2.5.2 specifies. A
block whose request identifier matches no outstanding call is discarded.

#### 3.2.6 Timer Events

None beyond section 3.1.6.

#### 3.2.7 Other Local Events

None.

### 3.3 Server Details

#### 3.3.1 Abstract Data Model

In addition to section 3.1.1:

- **PipeHandlers**: for each open pipe, the service bound to it.
- **SessionState**: per connection, the identity established by the service that
  authenticates it. Every pipe on the connection shares it. It is set after the
  pipe carrying the authentication is already open, so a service MUST NOT bind
  identity at pipe-open time.
- **OpenStreams**: for each pipe, the chunked fields being received, keyed by
  stream identifier.

#### 3.3.2 Timers

None beyond section 3.1.2.

#### 3.3.3 Initialization

The server performs steps 2 and 3 of section 3.1.3 and then answers control
frames per section 2.2.4.1.

#### 3.3.4 Higher-Layer Triggered Events

**Pushing a message.** A service may send a call block on a pipe at any time,
repeating the Class, Method, and RequestId of a call the client left outstanding
(section 3.2.5.3). Sends on one connection MUST be serialized, so that the
fragments of one message stay contiguous and sequence numbers stay monotonic.

#### 3.3.5 Message Processing Events and Sequencing Rules

##### 3.3.5.1 Receiving a Pipe Open Request

The server:

1. Sends the pipe-open response (section 2.2.4.3) in a Continuation frame.
2. Binds the named service to the pipe index. Matching on the service name is
   case-insensitive.
3. Sends the interface table for that service (section 2.2.7).

The interface table MUST be sent even when the service is not otherwise ready.
A client that opens a pipe and receives no table blocks until its own timeout
expires, then abandons the operation.

If the service name is unknown, the server answers the pipe-open request and
sends no table.

##### 3.3.5.2 Receiving a Call Block

1. If Class has bits 0xE0 set, process the block as a stream frame (section
   3.3.5.3).
2. If the body is the single byte 0x0F, answer the iterator cancel per section
   2.2.10.
3. Otherwise resolve Class to an interface and Method to a method, parse the
   body per section 2.2.8, and execute.
4. Build the reply body per section 2.2.9 and send it with the Class, Method,
   and RequestId of the call, unchanged.

For a method that returns nothing, no reply is sent. A method the server does
not implement MUST either return the error field of section 2.2.9.5 or send no
reply; returning a reply of a shape the receive descriptors did not ask for
fails the call at the client with a parameter type mismatch.

##### 3.3.5.3 Receiving a Stream Frame

The server appends the field bytes to the stream named by byte 1, and marks the
stream complete on class 0xE7. It MUST NOT reply.

Stream frames arrive **after** the call block that references them, and the
client does not wait for them to be consumed before sending its next call. A
method whose result depends on a chunked field MUST therefore answer the call
block when it arrives, and hold the operation that consumes the field until
every stream the call referenced has seen its 0xE7 frame.

##### 3.3.5.4 Receiving a Pipe Close

The server releases the service bound to the pipe, discards the pipe's
reassembly state, and discards any streams open on it. When every pipe a client
opened has been closed, the server closes the connection.

#### 3.3.6 Timer Events

None beyond section 3.1.6.

#### 3.3.7 Other Local Events

None.

## 4 Protocol Examples

Every byte sequence in this section is complete and consistent with the rules
above. Bytes are shown in hexadecimal, most significant nibble first.

### 4.1 Connection Establishment

Client to server, link initialization probe:

```
0d
```

Server to client:

```
43 4f 4d 0d                       "COM" CR
```

Server to client, transport parameters, sequence 0:

```
80 80 e0 17 00 ff ff 03 00 04 00 00 00 04 00 00
1b 32 00 00 00 01 00 00 00 58 02 00 00 38 c9 9a
7e 0d
```

| Bytes | Meaning |
|---|---|
| `80` | Seq: bit 7 set, sequence 0 |
| `80` | Ack: bit 7 set, expecting sequence 0 |
| `e0` | Pipe header: pipe 0, Continuation, LastData |
| `17 00` | ContentLength 23 |
| `ff ff` | Routing: control frame |
| `03` | Control type 3, transport parameters |
| `00 04 00 00` | PacketSize 1024 |
| `00 04 00 00` | MaxBytes 1024 |
| `1b 32` `00 00 00` | WindowSize 16. The value byte 0x10 is escape-encoded as `1b 32` |
| `01 00 00 00` | AckBehind 1 |
| `58 02 00 00` | AckTimeout 600 ms |
| `38 c9 9a 7e` | Check field |
| `0d` | Terminator |

Client to server, connection established:

```
80 80 e0 03 00 ff ff 04 fc 03 18 a0 0d
```

Server to client, the echo of a connection request whose body was
`01 00 00 00`. The client's own frame carries the same control type and the same
body, under its own sequence and acknowledgment numbers:

```
81 81 e0 07 00 ff ff 01 01 00 00 00 41 06 a0 ed 0d
```

| Bytes | Meaning |
|---|---|
| `e0` | Pipe header: pipe 0, Continuation, LastData |
| `07 00` | ContentLength 7 |
| `ff ff` | Routing: control frame |
| `01` | Control type 1, connection request |
| `01 00 00 00` | Body, repeated from the request unchanged |

Server to client, acknowledgment of sequence 0:

```
41 81 f2 cd dd 73 0d
```

| Bytes | Meaning |
|---|---|
| `41` | Acknowledgment packet |
| `81` | Ack: bit 7 set, expecting sequence 1 |
| `f2 cd dd 73` | Check field |
| `0d` | Terminator |

### 4.2 Opening a Service Pipe

Client to server, opening service `LOGSRV` version 6 on pipe 3:

```
81 81 e0 13 00 00 00 00 00 03 00 4c 4f 47 53 52
56 00 55 00 06 00 00 00 9b 26 95 0e 0d
```

| Bytes | Meaning |
|---|---|
| `81 81` | Seq 1, Ack 1 |
| `e0` | Pipe header: pipe 0, Continuation, LastData |
| `13 00` | ContentLength 19 |
| `00 00` | Routing: pipe-open request |
| `00 00` | Reserved |
| `03 00` | PipeIndex 3 |
| `4c 4f 47 53 52 56 00` | ServiceName "LOGSRV" |
| `55 00` | Parameter "U" |
| `06 00 00 00` | Version 6 |

Server to client, pipe-open response:

```
82 83 e3 08 00 03 00 01 00 03 00 00 00 b6 d3 09
2d 0d
```

| Bytes | Meaning |
|---|---|
| `82 83` | Seq 2, Ack 3 |
| `e3` | Pipe header: pipe 3, Continuation, LastData |
| `08 00` | ContentLength 8 |
| `03 00` | Routing: pipe 3 |
| `01 00` | Command: pipe opened |
| `03 00` | ServerPipeIndex 3 |
| `00 00` | Status: success |

Server to client, interface table. The host block and the first three of its ten
records:

```
00 00 00
b6 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46 01
b7 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46 02
b8 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46 03
```

| Bytes | Meaning |
|---|---|
| `00 00 00` | Class 0, Method 0, RequestId 0 |
| `b6 8b 02 00 …46` | GUID 00028BB6-0000-0000-C000-000000000046 |
| `01` | Interface identifier 1 for that GUID |

### 4.3 A Call with a Static Reply

Client to server, call block: interface 6, method 0, request identifier 0. The
body carries a dword, an 88-byte variable field, and eight receive descriptors.

```
06 00 00
03 43 16 00 00
04 d8
00 00 00 00 6d 69 63 72 6f 73 6f 66 74 00 00 00
00 00 00 00 00 00 00 00 00 00 00 00 00 00 01 00
00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
00 00 00 00 00 74 65 73 74 00 00 00 01 00 00 00
c0 32 3b 82 00 00 00 00
83 83 83 83 83 83 83 84
```

| Bytes | Meaning |
|---|---|
| `06` | Class: interface 6 |
| `00` | Method 0 |
| `00` | RequestId 0, one-byte VLI |
| `03 43 16 00 00` | Send parameter, dword, value 5699 |
| `04 d8` | Send parameter, variable, size byte 0xD8: bit 7 set, length 0x58 = 88 |
| `00 00 …` | The 88 bytes of the field |
| `83` × 7 | Seven receive descriptors, dword |
| `84` | One receive descriptor, variable |

Server to client, reply. Seven dwords, end-of-static, and a 16-byte variable
field, on pipe 3:

```
84 82 e3 3b 00 03 00 06 00 00 83 00 00 00 00 83
00 00 00 00 83 00 00 00 00 83 00 00 00 00 83 00
00 00 00 83 00 00 00 00 83 00 00 00 00 87 84 1b
35 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
00 d7 f0 ca 3f 0d
```

| Bytes | Meaning |
|---|---|
| `84 82` | Seq 4, Ack 2 |
| `e3` | Pipe header: pipe 3, Continuation, LastData |
| `3b 00` | ContentLength 59 |
| `03 00` | Routing: pipe 3 |
| `06 00 00` | Class 6, Method 0, RequestId 0, repeated from the call |
| `83 00 00 00 00` × 7 | Seven dword fields, answering the seven `83` descriptors |
| `87` | End of static section |
| `84 1b 35 …` | Variable field. The size byte 0x90 — bit 7 set, length 16 — is escape-encoded as `1b 35` |
| `d7 f0 ca 3f` | Check field |

### 4.4 A Call with a Dynamic Reply

A call that declares the descriptors `83 85` — one dword and one dynamic field —
is answered with a status dword, the end-of-static tag, and an opaque blob that
runs to the end of the host block:

```
83 00 00 00 00 87 86 de ad be ef
```

| Bytes | Meaning |
|---|---|
| `83 00 00 00 00` | Status dword |
| `87` | End of static section |
| `86` | Dynamic content follows; the call completes when it is consumed |
| `de ad be ef` | Content, to the end of the host block |

The same reply with `88` in place of `86` would deliver the content as one
message of a sequence and leave the call outstanding, and a caller waiting for a
single result would never be released.

A reply with an empty static section is legal and starts with the end-of-static
tag:

```
87 88 …content…
```

followed by a second host block, on the same Class, Method, and RequestId, that
ends the sequence:

```
87 86
```

### 4.5 A Reply Fragmented Across Packets

A reply whose pipe message is 2060 bytes, sent on pipe 5 with PacketSize 1024,
starting at sequence 10. Three packets carry it:

| Packet | Wire length | Seq | Pipe header | LastData | Payload bytes |
|---|---|---|---|---|---|
| 1 | 1024 | 10 | `a5` | no | 989 |
| 2 | 1024 | 11 | `a5` | no | 989 |
| 3 | 94 | 12 | `e5` | yes | 87 |

First bytes of packet 1:

```
8a 84 a5 0c 08 05 00 03 …
```

| Bytes | Meaning |
|---|---|
| `8a` | Seq 10 |
| `84` | Ack 4 |
| `a5` | Pipe header: pipe 5, Continuation, LastData clear |
| `0c 08` | ContentLength 2060, the length of the whole message |
| `05 00` | Routing: pipe 5 |
| `03 …` | Host block |

First bytes of packet 2:

```
4b 84 a5 ce …
```

| Bytes | Meaning |
|---|---|
| `4b` | Seq 11. The value `0x8b` is a reserved value and is transmitted exclusive-ORed with 0xC0 |
| `84` | Ack 4 |
| `a5` | Pipe header: pipe 5, Continuation, LastData clear |
| `ce …` | Message bytes. No length field: only frame 1 carries one |

Packet 1 contributes 986 message bytes — its 989 payload bytes less the header
and the two-byte length — packet 2 contributes 988, and packet 3 contributes 86.
The total reaches 2060 and the message is delivered.

### 4.6 A Call Carrying a Chunked Field

Client to server, a call on interface 11, method 4, request identifier 7, whose
second argument is a 1500-byte field sent out of band:

```
0b 04 07 01 04 05 01 dc 05 00 00 83 84
```

| Bytes | Meaning |
|---|---|
| `0b 04 07` | Class 11, Method 4, RequestId 7 |
| `01 04` | Send parameter, byte, value 4 |
| `05` | Send parameter, chunked reference |
| `01` | StreamId 1 |
| `dc 05 00 00` | Length 1500 |
| `83 84` | Receive descriptors: dword, variable |

The field then follows on the same pipe as stream frames, the last carrying
class 0xE7:

```
e6 01 41 41 41 41 41 41 41 41        eight bytes of the field
…
e7 01 42 42 42 42                    the final four bytes
```

The server answers the call block as soon as it arrives, and completes the
operation that consumes the field once the 0xE7 frame has been received.

## 5 Security

### 5.1 Security Considerations for Implementers

MOS RPC provides no confidentiality, no integrity protection against
modification, and no authentication of either peer. The check field detects
corruption, not tampering: it is a CRC with a published polynomial, and any
party that can modify the byte stream can recompute it.

Credentials carried as call parameters are visible to anyone with access to the
link. Deployments that need confidentiality must obtain it below this protocol.

Authentication is a service concern. Identity established by one service applies
to the whole connection (section 3.3.1), so a server MUST treat every pipe on a
connection as carrying the same principal, and MUST NOT infer identity from the
pipe a call arrives on.

### 5.2 Index of Security Parameters

An implementation MUST bound the following, all of which are attacker-controlled
in a message and each of which otherwise allows a peer to force unbounded
allocation or a non-terminating loop:

| Parameter | Bound |
|---|---|
| Packet length | PacketSize. Discard longer packets. |
| ContentLength | 65,535 bytes, and the buffer MUST be released if the pipe closes before the message completes. |
| Message reassembly | A frame that would carry more bytes than BytesOutstanding MUST be truncated to it, never allowed to overrun. |
| Variable field size | 32,767 bytes, and MUST NOT exceed the bytes remaining in the host block. |
| Dynamic message | 16,384 bytes per message. |
| Send parameters | 16 per request. |
| Receive descriptors | 16 per request. |
| Chunked field length | The declared length. Discard stream bytes beyond it. |
| Concurrent streams | One per stream identifier, 256 per connection. |
| Pipes | 16 per connection. |

## 6 Appendix A: Constants

**Packet framing**

| Name | Value |
|---|---|
| Terminator | 0x0D |
| Escape byte | 0x1B |
| Acknowledgment packet type | 0x41 |
| Negative acknowledgment packet type | 0x42 |
| Reserved values requiring the transform of section 2.2.2.2 | 0x8D, 0x90, 0x8B |
| Transform mask | 0xC0 |
| Check polynomial | 0x248EF9BE |
| Check mask | OR 0x60 on 0x1B, 0x0D, 0x10, 0x0B, 0x8D, 0x90, 0x8B |
| Minimum packet length | 7 bytes |

**Pipe header bits**

| Name | Mask |
|---|---|
| PipeIndex | 0x0F |
| HasLength | 0x10 |
| Continuation | 0x20 |
| LastData | 0x40 |
| Always set | 0x80 |

**Routing values**

| Name | Value |
|---|---|
| Pipe open | 0x0000 |
| Control | 0xFFFF |
| Pipe data | pipe index, 0x0001 - 0x000F |

**Control frame types**

| Type | Name |
|---|---|
| 1 | Connection request |
| 3 | Transport parameters |
| 4 | Connection established |

**Host block classes**

| Value | Meaning |
|---|---|
| 0x00 | Interface table |
| 0x01 - 0xDF | Interface identifier |
| 0xE6 | Stream frame, more follow |
| 0xE7 | Stream frame, last |

**Request tags**

| Tag | Meaning |
|---|---|
| 0x01 | byte |
| 0x02 | word |
| 0x03 | dword |
| 0x04 | variable |
| 0x05, 0x45 | chunked field reference |
| 0x81 - 0x85 | receive descriptors |

**Reply tags**

| Tag | Meaning |
|---|---|
| 0x81 | byte |
| 0x82 | word |
| 0x83 | dword |
| 0x84 | variable |
| 0x85 | dynamic, message stays open |
| 0x86 | dynamic, call complete |
| 0x87 | end of static section |
| 0x88 | dynamic, message complete |
| 0x8F | error |

**Error values**

See section 2.2.9.5.

**Defaults and negotiated limits**

| Name | Default | Source |
|---|---|---|
| PacketSize | 1024 | Negotiated |
| MaxBytes | 1024 | Negotiated |
| WindowSize | 16 | Negotiated |
| AckBehind | 1 | Negotiated |
| AckTimeout | 600 ms | Negotiated |
| KeepAlive | none | Negotiated, optional |
| Retransmission limit | 12 | Fixed |
| Sequence space | 128 | Fixed |
| Pipes per connection | 16 | Fixed |
| Dynamic message capacity | 16,384 bytes | Fixed |
| Link initialization timeout | 20 s | Fixed |

## 7 Appendix B: Behavior Notes

**Pipe index in the frame header.** Implementations differ in which value they
put in the PipeIndex field of a frame carrying data for pipe *n*. One sends
every message with PipeIndex 0 and relies on the routing value; the other sets
PipeIndex to *n* and sends the routing value as well. Both forms are legal, and
a receiver that routes on the routing value (section 3.1.5.5) interoperates with
either.

**Escape encoding of frame headers.** A pipe header or length byte whose value
is one of the reserved values is transformed per section 2.2.2.2 and therefore
never reaches the escape encoder. A receiver that reverses the transform on
these bytes and unescapes the rest of the payload handles both that form and a
sender that leaves the value in place.

**The `01` VLI form.** Bits 7-6 of `01` decode identically to `00`. Senders emit
`00`.

**Reserved trailing descriptor.** Requests commonly end with a `84` receive
descriptor that the method itself does not define. It is answered like any other
variable descriptor, after the end-of-static tag.

**LastData and message completion.** Some senders leave LastData clear on the
frame that completes a message and set it only on the last frame of a burst.
Receivers that end a message on LastData rather than on ContentLength
concatenate two messages and hand the application a buffer with a frame header
embedded in it.

**Unknown service names.** A server that does not recognize a service name still
answers the pipe-open request. The pipe opens and carries no interface table;
the client abandons it after its own timeout.

**Unimplemented methods.** A server that neither replies nor returns an error
leaves the call outstanding indefinitely. Clients that expose no call timeout of
their own hang until the pipe closes.
