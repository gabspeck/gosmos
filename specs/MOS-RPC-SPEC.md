# MOS Remote Procedure Call Protocol Specification

## 1 Introduction

The MOS Remote Procedure Call Protocol (MOS RPC) carries remote procedure callsmos
between a client application and an online service. It addresses calls to an
interface and a method with typed parameters, multiplexes up to 16 independent
logical channels onto one connection, and streams bulk arguments and result sets
out of band.

MOS RPC does not frame its own bytes. It rides on a **transport binding**, which
carries pipe messages between the two peers in order and without loss. Two
bindings are defined:

| Binding | Link | Machinery |
|---|---|---|
| Select | serial | packets, escape encoding, check field, sliding window, fragmentation |
| Straight | TCP | length-prefixed records |

Sections 2 and 3 define the protocol, and hold for every binding. Section 4
defines the two bindings. A peer implementing the protocol and one binding
interoperates with a peer implementing the protocol and the same binding;
nothing above section 4 changes when the binding changes.

The protocol is organized in two layers:

- **Pipe layer.** Multiplexes 15 logical channels (pipes) onto one connection,
  each carrying one service. Connection control and pipe open ride reserved
  routing values rather than a pipe of their own.
- **Call layer.** Carries requests and replies as host blocks addressed to an
  interface and a method, with typed parameters.

### 1.1 Glossary

Terms marked *(Select)* or *(Straight)* belong to a transport binding and appear
only in section 4.

**call** — a request sent by a client on a pipe, together with the reply the
server returns for it.

**call block** — a host block that carries a call or a reply.

**check field** — the four-byte error-detection field at the end of a packet.
*(Select)*

**chunked field** — a variable-length parameter too large to travel inside a
call block, sent instead as a separate sequence of stream frames.

**class** — the first byte of a host block, which selects the layout of the rest
(section 2.2.5). On a call block it is the interface identifier.

**client** — the peer that establishes the connection, opens pipes, and issues
calls.

**connection** — one instance of the protocol running over one transport
binding.

**dynamic message** — one message assembled from the dynamic sections of one or
more host blocks on a single request identifier, and delivered to the caller
when a 0x88 or 0x86 tag closes it (section 2.2.8.4).

**dynamic section** — the trailing part of a reply body whose content is opaque
to the call layer and runs to the end of the host block.

**host block** — the unit of the call layer: a class byte and a body whose
layout the class selects (section 2.2.5).

**interface** — a named set of methods, identified by a GUID and addressed on
the wire by a one-byte interface identifier assigned by the server.

**interface table** — the message that maps interface GUIDs to interface
identifiers on a newly opened pipe.

**method** — one operation of an interface, addressed by a one-byte method
identifier.

**packet** — the unit of the Select binding, terminated by the byte 0x0D.
*(Select)*

**peer** — either party to a connection.

**pipe** — one of 15 logical channels multiplexed on a connection, numbered 1
through 15.

**pipe frame** — the unit into which the Select binding cuts a pipe message: a
header byte, an optional length byte, and content. *(Select)*

**pipe index** — the number 1 through 15 that names a pipe.

**pipe message** — the unit of the pipe layer: a routing value followed by
content.

**pipe unit** — the unit a transport binding carries: one pipe message, plus the
framing field the binding wraps around it (section 2.1.1).

**receive descriptor** — a tag in a request body that declares the type of one
field the caller expects in the reply.

**reassembly index** — the number 0 through 15 the Select binding gives a pipe
message while it is in flight, so that a receiver can tell partially received
messages apart. It is not a pipe index (section 4.2.3). *(Select)*

**record** — the unit of the Straight binding: a length, a command-byte echo,
and one pipe message. *(Straight)*

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

**transport binding** — the framing and reliability machinery that carries pipe
messages over one kind of link. Section 4.

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

A connection begins when the transport binding establishes its link. The client
announces itself, the server sends its transport parameters, the client sends a
connection request and the server echoes it (section 3.1.3). The two then
exchange pipe messages until one of them tears the connection down.

Connection control — transport parameters, the connection request and its
confirmation — travels on the reserved routing value 0xFFFF, and a pipe-open
request on the reserved routing value 0x0000. Neither is a pipe, and no message
ever names pipe 0: the number is not addressable, its routing value being the
pipe-open request itself. To reach a service, the client sends a pipe-open
request naming the service, a version, and the pipe index it intends to use.
The server answers on that index and immediately sends the interface table for
the service, mapping each interface GUID the service supports to a one-byte
interface identifier.

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
fields: the call block carries a reference — a tag, a stream identifier, and a
length, six bytes in all — and the bytes follow in one-way stream frames on the
same pipe.

```
  application        calls, replies, notifications
  ------------------------------------------------------------------
  call layer         host blocks, interfaces, methods, parameters
  ------------------------------------------------------------------
  pipe layer         15 logical channels, routing, open and close
  ==================================================================
  transport binding  Select                    | Straight
                     packets, escape encoding, | length-prefixed
                     check field, window,      | records
                     frames, reassembly        |
  ------------------------------------------------------------------
  link               serial link               | TCP connection
```

### 1.4 Relationship to Other Protocols

MOS RPC requires a transport binding that meets section 2.1.1. The Select
binding builds that service out of an ordered byte stream, taking nothing from
the link beyond the bytes. The Straight binding takes it from TCP.

Services carried over MOS RPC define their own message contracts on top of the
call layer. Those contracts are out of scope for this document, which specifies
the framing, addressing, and parameter encoding every one of them uses.

### 1.5 Prerequisites/Preconditions

A link to the server endpoint must already exist. The endpoint selects the
transport binding (section 4.1), and both peers must be using the same one.
Dialling, connecting, and any modem or bridge in between are outside this
specification; section 4 states what each binding does once the link is up.

### 1.6 Applicability Statement

MOS RPC suits request-response and streaming interaction between one client and
one server. Over the Select binding its window, packet size, and per-pipe
reassembly are sized for links of a few kilobytes per second. It provides no
confidentiality or integrity guarantee against an active attacker on any
binding; see section 6.

### 1.7 Versioning and Capability Negotiation

The protocol negotiates in three places:

- **Transport parameters.** The server sends its packet size, window, and timer
  values (section 2.2.3.1.2), which govern the Select binding; Straight ignores
  them but the client still requires the message. Section 2.2.3.1.2 states how
  each peer applies them.
- **Service version.** The pipe-open request names a service version (section
  2.2.3.2). There is no encoding for refusing an open: Command and Status carry
  one defined value each (section 2.2.3.3), so a server that will not serve the
  request answers it and withholds the interface table (section 3.3.5.1).
- **Interfaces.** The interface table (section 2.2.6) tells the client which
  interfaces the service exposes on this connection and what identifier each has
  on this pipe. Identifiers are per-pipe assignments, not fixed constants; a
  client MUST resolve them from the table on every pipe it opens.

There is no protocol version number on the wire, and no binding identifier: the
peers agree on the binding out of band, by the link they connected over.

### 1.8 Vendor-Extensible Fields

Interface identifiers 0x01 through 0xDF are assigned by the server per pipe and
carry no meaning outside a connection. Every value in that range is available:
this specification reserves none of them, and a server MAY assign them in any
order it likes. Services extend the protocol by defining new interface GUIDs and
publishing them in the interface table.

The class range 0xE0 through 0xFF is reserved by this specification (section
2.2.5). The tag values in sections 2.2.7 and 2.2.8 are reserved; a receiver that
meets an unassigned send-parameter tag cannot determine the length of the field
that follows it and MUST stop parsing the body at that point. An unassigned
receive-descriptor tag carries no field and is ignored instead (section
2.2.7.3).

### 1.9 Standards Assignments

None. Interface identity uses GUIDs, which require no assignment authority.

## 2 Messages

### 2.1 Transport

#### 2.1.1 Abstract Transport Service

A transport binding carries **pipe units** between the two peers. A pipe unit is
one pipe message (section 2.2.2). Each binding wraps it in a framing field of
its own — a reassembly index on Select (section 4.2.3), an echo of the message's
command byte on Straight (section 4.3.1) — and neither is addressing (section
3.1.5.1).

A binding MUST provide:

| Property | Requirement | Discharged by |
|---|---|---|
| Boundaries | Each unit is delivered whole and separate. The receiver never sees two messages joined or one split. | Select: ContentLength (4.2.2, 4.2.9). Straight: TotalLength (4.3.1, 4.3.5). |
| Order | Units are delivered in the order they were submitted, on each pipe and across pipes. | Select: sequence numbers (4.2.7). Straight: TCP. |
| Atomicity | A peer MUST NOT interleave the frames of two pipe units that share a reassembly index. Frames on distinct indexes may interleave freely. | Select: one reassembly context per index (4.2.3). Straight: nothing to interleave, a record is always whole (4.3.1). |
| Reliability | Every unit submitted is delivered exactly once, or the connection fails. | Select: retransmission and duplicate suppression (4.2.6, 4.2.7, 4.2.10). Straight: TCP. |
| Bidirectionality | Either peer may submit a unit at any time once the binding reports the connection up. Bring-up order (section 3.1.3) constrains only bring-up. | Select: 4.2. Straight: TCP, full duplex. |
| Capacity | At least 65,532 bytes of pipe message per unit. | Select: fragmentation (4.2.8). Straight: TotalLength (4.3.1). |
| Failure | Loss of the link is reported to the pipe layer as one event. | Both: 4.1. |
| Discard | On connection failure every partial message is discarded. A binding holds no per-pipe state and cannot be told to discard by pipe: it does not know which reassembly index, if any, a closing pipe's traffic was using. | Select: the reassembly table (4.2.3). Straight: nothing to discard. |

Only Select has anything to reassemble, and its reassembly index is what keeps
that state finite. A binding may not read the routing value, so the index is the
only thing that tells two partially received messages apart, and sixteen of them
cap what a receiver holds at once. Two messages sharing one
index cannot be separated: a peer that interleaves them destroys both, and the
receiver has no diagnostic — the joined bytes are delivered to the pipe layer as
one message and discarded there for an unrecognized routing value.

A binding provides nothing else. It does not interpret the message, does not
route on the routing value, and does not know what a pipe carries.

#### 2.1.2 Defined Bindings

| Binding | Section | Link |
|---|---|---|
| Select | 4.2 | serial |
| Straight | 4.3 | TCP |

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

| Bits 7-6 of byte 0 | Size | Value | Decoded range |
|---|---|---|---|
| 00 | 1 byte | `b0 & 0x3F` | 0 - 63 |
| 01 | — | — | Invalid |
| 10 | 2 bytes | `((b0 & 0x3F) << 8) \| b1` | 0 - 16,383 |
| 11 | 4 bytes | `(b0..b3 as big-endian uint32) & 0x3FFFFFFF` | 0 - 1,073,741,823 |

A receiver MUST reject `01` in bits 7-6 and discard the whole host block that
carries it. The form has no defined length, so a receiver cannot skip the field
to reach the rest of the block.

A sender selects the form by strict comparison against each form's ceiling, so
the boundary value travels in the next form up:

| Value | Form |
|---|---|
| 0 - 62 | 1 byte |
| 63 - 16,382 | 2 bytes |
| 16,383 - 1,073,741,822 | 4 bytes |
| 1,073,741,823 | Not encodable |

63 therefore goes out as `80 3f` and 16,383 as `c0 00 3f ff`. 0x3FFFFFFF has no
encoding at all: a sender asked to encode it fails the call rather than emit a
VLI. A receiver MUST accept any form wide enough to hold the value — `80 03` for
the value 3 — rather than reject the block: the value is unambiguous, every
valid form has a known length, and the non-shortest forms are what a sender
produces at the boundaries.

Examples:

| Value | Encoding |
|---|---|
| 0 | `00` |
| 62 | `3e` |
| 63 | `80 3f` |
| 16,382 | `bf fe` |
| 16,383 | `c0 00 3f ff` |
| 1,073,741,822 | `ff ff ff fe` |

##### 2.2.1.3 Variable-Length Field Size

Variable-length parameter fields (sections 2.2.7.1 and 2.2.8.3) prefix their
content with a size in one of two forms, selected by bit 7 of the first byte.

| Bit 7 of byte 0 | Size | Value | Decoded range |
|---|---|---|---|
| 1 | 1 byte | `b0 & 0x7F` | 0 - 127 |
| 0 | 2 bytes | `(b0 << 8) \| b1`, big-endian | 0 - 32,767 |

Note that bit 7 selects the **short** form here, the opposite polarity to the
VLI of section 2.2.1.2, where bits 7-6 of `00` select the short form. The
inversion buys the short form a seventh value bit, covering the 64 - 127 band
that carries most string arguments in one byte, and leaves the long form a plain
big-endian 15-bit integer that needs no masking to read. The two encodings are
not interchangeable and appear in different fields.

A sender selects the form by strict comparison, as in section 2.2.1.2:

| Length | Form |
|---|---|
| 0 - 126 | 1 byte |
| 127 - 32,766 | 2 bytes |
| 32,767 and above | No inline encoding |

A receiver MUST accept the two-byte form for a length below 128. The two forms
are distinguished by bit 7 alone, so `00 7f` is unambiguous: a one-byte size
always has bit 7 set, and a first byte with bit 7 clear can only begin a
two-byte size. Rejecting it would fail the call over the encoding a conforming
sender uses for the length 127.

Examples:

| Value | Encoding | Form |
|---|---|---|
| 0 | `80` | one byte |
| 16 | `90` | one byte |
| 88 | `d8` | one byte |
| 126 | `fe` | one byte |
| 127 | `00 7f` | two bytes |
| 300 | `01 2c` | two bytes |
| 32,766 | `7f fe` | two bytes |

A field of 32,767 bytes or more MUST be sent as a chunked field (section
2.2.7.2). A sender MAY send a shorter field as a chunked field too; section
2.2.7.2 states when.

##### 2.2.1.4 Interface Identifier GUID

A GUID travels as 16 bytes in structure layout: a 32-bit field, two 16-bit
fields, and eight bytes, where the first three fields are little-endian and the
final eight bytes are in order.

The GUID `00028BB6-0000-0000-C000-000000000046` encodes as:

```
b6 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46
```

Receivers compare GUIDs byte for byte.

#### 2.2.2 Pipe Message

```
+---------------+-------------------------------+
| RoutingValue  | Content                       |
+---------------+-------------------------------+
        2                  0 .. 65530
```

| Field | Size | Description |
|---|---|---|
| RoutingValue | 2 | Little-endian. Selects the kind and destination of the message. |
| Content | variable | Determined by the routing value. |

| Routing value | Content |
|---|---|
| 0x0000 | Pipe-open request (section 2.2.3.2) |
| 0x0001 - 0x000F | Data for that pipe (section 2.2.3.4) |
| 0x0010 - 0xFFFE | Unassigned. A receiver MUST discard the message. |
| 0xFFFF | Control frame (section 2.2.3.1) |

Every pipe message begins with a routing value, whatever it is for and in both
directions. A receiver dispatches on it per section 3.1.5.1.

A message whose routing value is a pipe index carries the pipe-open response
(section 2.2.3.3) when the pipe is still opening, and pipe data (section
2.2.3.4) once it is open. Nothing in the message distinguishes the two; the
receiver knows which it is from the state of the pipe.

A pipe message MUST NOT exceed 65,532 bytes, the least any binding carries
(section 2.1.1). The figure comes from the Straight binding, whose 16-bit
TotalLength counts a 3-byte record header along with the message: 65,535 − 3.
Payloads larger than that are delivered as a sequence of host blocks on one
request identifier (section 2.2.8.4), never as one message. On the
Select binding ContentLength can express 65,535; a receiver MUST discard a
message whose declared length exceeds 65,532 rather than reassemble it (section
6.2).

#### 2.2.3 Pipe Message Content

##### 2.2.3.1 Control Frame

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

Types 2 and 5 through 255 are unassigned. A receiver MUST discard a control
frame of an unassigned type and MUST NOT treat it as an error: the connection
continues, and bring-up is unaffected. A control frame carrying no type byte at
all — a two-byte pipe message, the routing value alone — is discarded on the
same rule.

###### 2.2.3.1.1 Connection Request (Type 1)

```
+--------------+--------+-----------------+
| 0xFFFF       |  0x01  | Type-specific   |
+--------------+--------+-----------------+
       2           1          4 .. N
```

The type-specific field is opaque to the server, which MUST return a type 1
control frame whose type-specific field repeats the received bytes exactly, and
MUST NOT parse it.

The field is four bytes to several hundred, and a receiver MUST bound what it
buffers to echo (section 6.2). A frame whose type-specific field exceeds that
bound is discarded and not echoed; the client's bring-up then stalls on its own
timeout, which is the same outcome as a lost frame.

The client uses the field to carry a connection log — a modem description and
the addresses it failed to reach, as `host!err|count|MMDDYYHHMM` records — which
has no protocol meaning.

###### 2.2.3.1.2 Transport Parameters (Type 3)

```
+--------+------+-----------+-----------+-----------+-----------+-----------+-----------+
| 0xFFFF | 0x03 | PacketSize| MaxBytes  | WindowSize| AckBehind | AckTimeout| KeepAlive |
+--------+------+-----------+-----------+-----------+-----------+-----------+-----------+
    2       1         4           4           4           4           4        0 or 4
```

The six parameters are little-endian 32-bit unsigned integers. Their meaning is
defined by the Select binding (section 4.2); this section defines only the
layout.

| Field | Description |
|---|---|
| Routing | 0xFFFF. |
| Type | 0x03. |
| PacketSize | Maximum Select packet length in bytes (section 4.2.1). |
| MaxBytes | Maximum bytes the sender will hold outstanding. |
| WindowSize | Maximum unacknowledged packets in flight. |
| AckBehind | Number of unacknowledged received packets after which the receiver MUST acknowledge. |
| AckTimeout | Retransmission timer, in milliseconds. |
| KeepAlive | Idle interval in milliseconds. Optional. |

The message carries either five fields or six. KeepAlive is present only when
the pipe message is 27 bytes long; a 23-byte message carries none. No other
length is defined, and a receiver MUST discard a message of any other length.

**Applying the values.** Each value is a limit on what its **sender** will
emit. A peer applies the smaller of each received value and its own configured
maximum to what it emits itself, and MUST accept any packet up to the PacketSize
its peer advertised. Only the server sends this message, so only the server's
limits are ever announced; a client whose own maximum is lower reduces what it
sends and does not reduce what it accepts. A peer that discarded packets larger
than its own configured maximum would silently kill a connection in which
nothing was wrong, because no message carries the client's values back and the
server cannot learn them.

**Defaults.** A server that has no configured value for a parameter sends these:

| Name | Default |
|---|---|
| PacketSize | 1024 |
| MaxBytes | 1024 |
| WindowSize | 16 |
| AckBehind | 1 |
| AckTimeout | 600 ms |
| KeepAlive | none — the five-field message |

The values are otherwise the server's own choice, bounded by section 6.2.

Every binding MUST carry this message, and the server MUST send it as its first
pipe message (section 3.1.3), even when the binding makes no use of the values.
A client does not proceed past bring-up until it arrives.

###### 2.2.3.1.3 Connection Established (Type 4)

```
+--------------+--------+
| 0xFFFF       |  0x04  |
+--------------+--------+
       2           1
```

The type-specific field is empty. The client sends this frame as soon as the
binding's framing is up and before the transport parameters reach it, on every
binding.

##### 2.2.3.2 Pipe Open Request

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

Both strings are bounded by section 6.2. A request in which either string runs
to the end of the message without a NUL terminator, in which either exceeds its
bound, or which has fewer than four bytes left for Version, is malformed and is
discarded per section 3.3.5.1.

##### 2.2.3.3 Pipe Open Response

```
+-----------+---------+----------------+--------+
| PipeIndex | Command | ServerPipeIndex| Status |
+-----------+---------+----------------+--------+
      2          2           2              2
```

| Field | Size | Description |
|---|---|---|
| PipeIndex | 2 | Routing value: the pipe index from the request. |
| Command | 2 | 0x0001, pipe opened. The only defined value. |
| ServerPipeIndex | 2 | The index the server will use. Equal to PipeIndex. |
| Status | 2 | 0x0000. No other value is defined. |

The message is eight bytes. Its first field is the routing value, and a sender
MUST NOT place a further routing value in front of it: the eight bytes are the
whole message. A client that receives a ten-byte response with the routing value
repeated stops there and never opens the pipe.

This message MUST be the first the server sends on a pipe it has just opened,
and it is what moves the pipe from opening to open. A receiver reads a message
on a pipe in the opening state as this response, and every message on that pipe
afterwards as pipe data (section 2.2.3.4); nothing in the bytes distinguishes
the two, so a receiver that does not make the transition on this message
misparses everything that follows.

##### 2.2.3.4 Pipe Data

```
+-----------+-------------------------+
| PipeIndex | Host block              |
+-----------+-------------------------+
      2             variable
```

The routing value is the destination pipe index, and the rest of the pipe
message is one host block (section 2.2.5).

#### 2.2.4 Pipe Close

```
+-----------+--------+
| PipeIndex |  0x01  |
+-----------+--------+
      2         1
```

A pipe message whose content after the routing value is the single byte 0x01
closes that pipe. No response is sent. Either peer may close a pipe it opened or
serves; after a close, both peers discard the pipe's state and any calls
outstanding on it. The binding is told nothing: its reassembly state is keyed on
reassembly indexes, not pipes (section 2.1.1).

A three-byte pipe message whose content is any single byte other than 0x01 is
unassigned and is discarded.

The close cannot be confused with pipe data, which begins with a Class byte:
0x01 is a legal interface identifier, but the shortest call block is three bytes
(Class, Method, and a one-byte RequestId), so the shortest pipe-data message is
five bytes and a three-byte message is never pipe data.

#### 2.2.5 Host Block

```
+--------+----------------------------+
| Class  | Class-specific             |
+--------+----------------------------+
    1              variable
```

The first byte of a host block is the Class, which selects the layout of the
rest.

| Class | Layout |
|---|---|
| 0x00 | Interface table (section 2.2.6) |
| 0x01 - 0xDF | Call block (section 2.2.5.1) |
| 0xE0 - 0xFF | Stream frame (section 2.2.5.2) |

##### 2.2.5.1 Call Block

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
| Body | variable | Request parameters (section 2.2.7) or reply parameters (section 2.2.8). |

A reply MUST repeat the Class, Method, and RequestId of the call it answers. A
client discards any reply that does not match an outstanding call in all three,
silently and without failing the call (section 3.2.5.2). A server that alters
any of the three — including a server that replies with a Class the interface
table for that pipe never published — sends a block the client throws away, and
the call stays outstanding until the pipe closes.

A RequestId identifies one call for as long as that call is outstanding. A
sender MUST NOT reuse one that is still outstanding on the pipe; a receiver that
meets a second call on a live RequestId MUST discard the second call rather than
overwrite the first, whose caller would otherwise be answered with the wrong
result.

##### 2.2.5.2 Stream Frame

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

A class in the range 0xE0 through 0xFF other than 0xE6 or 0xE7 is unassigned. A
receiver MUST discard such a host block rather than treat it as stream content,
which would corrupt the field it is appended to without any diagnostic.

A stream frame carries no method and no request identifier: byte 1 is the stream
identifier and the field bytes begin at byte 2. It is one-way. A receiver MUST
NOT reply to it; a reply would carry a request identifier the peer has no
outstanding call for and would be discarded, and the sender does not wait for
one.

#### 2.2.6 Interface Table

```
+--------+--------+--------+---------------------------------+
|  0x00  |  0x00  |  0x00  | Record [1..n]                   |
+--------+--------+--------+---------------------------------+
    1        1        1                 17 * n
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

Which GUIDs a table carries, how many records it holds, and which identifier
each GUID receives are all service-defined and server-chosen. This
specification fixes the record layout, the identifier range 0x01 - 0xDF, and
uniqueness within one table; it fixes nothing else, and a conformance test can
assert nothing else. A GUID MUST NOT appear twice in one table. The same service
may be bound to more than one pipe on a connection, and each such pipe receives
its own table; the identifiers in them need not agree.

#### 2.2.7 Request Parameters

A request body is a sequence of tagged fields: send parameters first, then
receive descriptors. A receiver MUST also accept the two interleaved.

Bit 7 of a tag distinguishes the two: clear for a send parameter, set for a
receive descriptor. Bits 3-0 give the type.

Bits 6-4 are not a field, and a receiver MUST NOT mask them off before matching:
the assigned tags are exactly the values listed in sections 2.2.7.1 and 2.2.7.3,
and every other value is unassigned and handled per section 1.8. Two tags set a
bit in that range: 0x44 and 0x45, on which bit 6 means the field content is
compressed (section 2.2.7.4). It does not generalize to any other tag — 0x11,
0x15, 0x25, 0x41, 0x42 and 0x43 are unassigned — and a receiver MUST match the
whole byte rather than test bit 6 on its own.

##### 2.2.7.1 Send Parameters

```
+--------+---------------------------+
| Tag    | Field                     |
+--------+---------------------------+
    1        per the table below
```

A 0x04 or 0x44 field carries its own size ahead of its content:

```
+--------+---------------+---------------+
|  0x04  | Size          | Content       |
+--------+---------------+---------------+
    1        1 or 2         0 .. 32766
```

Size uses the encoding of section 2.2.1.3.

The Field column gives what follows the tag; the tag byte is not counted in it.

| Tag | Type | Field |
|---|---|---|
| 0x01 | byte | 1 byte |
| 0x02 | word | 2 bytes, little-endian |
| 0x03 | dword | 4 bytes, little-endian |
| 0x04 | variable | size (section 2.2.1.3) followed by that many bytes |
| 0x44 | variable, compressed | size (section 2.2.1.3) followed by that many bytes, section 2.2.7.4 |
| 0x05 | chunked reference | 5 bytes, section 2.2.7.2 |
| 0x45 | chunked reference, compressed | 5 bytes, section 2.2.7.2, whose stream carries compressed content per section 2.2.7.4 |

These seven values are the whole of the assigned send-parameter tags. A tag
whose low nibble is 1 through 5 but which is not one of them — 0x11, 0x15, 0x25,
0x41, 0x42, 0x43 — is unassigned: a receiver MUST stop parsing the body at it
per section 1.8 and answer 0xE0000007 per section 3.3.5.2.

0x44 and 0x45 are not second encodings of 0x04 and 0x05. They share their
layout and differ in what the content is, so a receiver that treats them as
synonyms delivers compressed bytes to a method expecting the argument.

At most 16 send parameters may appear in one request (section 6.2).

##### 2.2.7.2 Chunked Field Reference

A variable-length argument that does not fit in the call block is replaced in
the body by a reference, and its bytes follow in stream frames (section
2.2.5.2).

A sender chooses the chunked form whenever the argument does not fit comfortably
in the room left in the call block. That decision is relative to the link's
packet size, not to the inline ceiling of section 2.2.1.3: on a link with
PacketSize 1024, a field of a few hundred bytes is already sent this way. A
receiver MUST NOT treat chunked fields as a rare path reserved for very large
arguments. The only absolute rule is that a field of 32,767 bytes or more has
no inline encoding and MUST be chunked.

```
+--------+----------+---------------+
| Tag    | StreamId | Length        |
+--------+----------+---------------+
    1         1            4
```

The reference is six bytes on the wire, of which five follow the tag:

| Field | Size | Description |
|---|---|---|
| Tag | 1 | 0x05, or 0x45 when the stream carries compressed content (section 2.2.7.4). |
| StreamId | 1 | Identifies the stream frames that carry this field. Unique among the streams outstanding on a connection. |
| Length | 4 | Total number of bytes the stream frames will carry, little-endian. On 0x45 that is the compressed count, not the size of the argument. |

Stream identifiers are allocated from a per-connection counter starting at 1, so
concurrent calls on one connection never share one. The namespace is per
connection on both sides: a receiver keys its open streams by identifier alone
(section 3.3.1), and a frame for a stream is accepted only on the pipe whose
call referenced it. A frame naming a live stream that arrives on another pipe is
discarded.

A reference naming a stream identifier that is already open is rejected: the
receiver answers the call with the error field 0xE0000009 and leaves the running
stream alone. Accepting it would merge two fields into one buffer, and a
conforming sender never produces one.

Length is bounded by section 6.2. A reference declaring more than that bound is
answered with the error field 0xE0000009 and its stream is never opened; the
frames that follow name no open stream and are discarded (section 3.3.5.3).

Section 5.6 shows a reference, its frames, and the reply.

A receiver MUST NOT parse the five bytes after a 0x05 or 0x45 tag as a
variable-length field. The StreamId byte would be read as a size and the rest of
the request consumed as content.

##### 2.2.7.3 Receive Descriptors

```
+--------+
| Tag    |
+--------+
    1
```

A receive descriptor is a tag with bit 7 set and no data. It declares the type
of one field the caller expects in the reply, in order.

| Tag | Declares |
|---|---|
| 0x81 | byte |
| 0x82 | word |
| 0x83 | dword |
| 0x84 | variable |
| 0x85 | dynamic |

At most 16 receive descriptors may appear in one request (section 6.2). The
reply's fields correspond to them positionally (section 2.2.8), except in an
error reply, which carries the error field alone (section 2.2.8.5).

A request may end with a 0x84 descriptor beyond the fields its method defines.
It is answered as a post-static field, after the end-of-static tag (section
2.2.8.3).

A receiver MUST treat every tag with bit 7 set as a receive descriptor and MUST
ignore one whose type it does not recognize, rather than stopping the parse:
section 1.8's rule applies to send-parameter tags, whose field length cannot be
determined, and a descriptor carries no field.

##### 2.2.7.4 Compressed Fields

A variable-length argument may be compressed by the sender. Bit 6 of the tag
says so: 0x44 is the compressed form of 0x04, and 0x45 of 0x05. Nothing else
about the encoding changes — a 0x44 field carries the size encoding of section
2.2.1.3 and a 0x45 reference the six-byte layout of section 2.2.7.2.

Compression happens before the sender chooses between the inline and chunked
forms, so **every length on the wire is the compressed length**: the size prefix
of a 0x44 field, and the Length of a 0x45 reference. The size of the argument
the caller supplied is not transmitted, and a receiver learns it only by
decompressing. A receiver decompresses until the input is consumed.

The compressed content is a sequence of blocks:

```
+---------------+-------------------+---------------+-------------------+
| BlockLength   | Block             | BlockLength   | Block             |
+---------------+-------------------+---------------+-------------------+
       4              variable             4              variable
```

| Field | Size | Description |
|---|---|---|
| BlockLength | 4 | Compressed length of the block that follows, little-endian. |
| Block | variable | One compressed block, expanding to at most 32,768 bytes. |

Blocks run to the end of the field; no count precedes them and no terminator
follows. Every block but the last expands to exactly 32,768 bytes.

**The block codec is not specified here.** It is a dictionary coder over a
32,768-byte window, and this document does not define its bitstream. Two
implementations therefore interoperate on the compressed forms only if they
agree on that codec out of band. A sender that cannot rely on the receiver
supporting it MUST use 0x04 and 0x05, which are always available and which no
receiver may refuse; sending 0x44 or 0x45 is never required of any sender.

A receiver that does not implement the codec MUST answer 0xE0000007 for an
inline 0x44 field and 0xE0000009 for a 0x45 reference, rather than deliver the
compressed bytes to the method as though they were the argument.

Compression is a request-side facility. No reply tag carries bit 6, and this
specification defines no compressed reply field.

#### 2.2.8 Reply Parameters

A reply body is a static section, an end-of-static tag, zero or more post-static
variable fields, and an optional dynamic section.

```
+----------------+--------+---------------------+-------------------+
| Static fields  |  0x87  | Post-static fields  | Dynamic section   |
+----------------+--------+---------------------+-------------------+
    variable          1          variable              variable
```

Everything after the static section is optional. A reply that ends at the
end-of-static tag is complete (section 2.2.8.2).

##### 2.2.8.1 Static Section

```
+--------+---------------------------+
| Tag    | Field                     |
+--------+---------------------------+
    1        per the table below
```

The static section holds one field per receive descriptor of the corresponding
type, in the order the descriptors appeared.

| Tag | Type | Field |
|---|---|---|
| 0x81 | byte | 1 byte |
| 0x82 | word | 2 bytes, little-endian |
| 0x83 | dword | 4 bytes, little-endian |
| 0x84 | variable | size (section 2.2.1.3) followed by that many bytes |
| 0x8F | error | 4 bytes, little-endian (section 2.2.8.5) |

The static section may be empty.

Only the first host block of a sequence carries a static section. Every later
block on the same request identifier MUST begin with the end-of-static tag,
whatever the call declared, so the two-byte bodies of sections 2.2.8.4 and 2.2.9
are well formed however many descriptors there were. A sender that repeated the
static fields on every block of a sequence would deliver them to the caller
again on each one.

##### 2.2.8.2 End-of-Static

```
+--------+
|  0x87  |
+--------+
    1
```

The tag 0x87 has no field and ends the static section. It MUST be present in
every reply, including replies whose static section is empty and replies that
carry no dynamic section.

A reply that ends at this tag carries no dynamic section and completes the call:
a caller waiting for a single result is released by it exactly as by 0x86. This
is the shape of an ordinary reply to a call that declared only fixed-size and
variable receive descriptors.

##### 2.2.8.3 Variable Fields

```
+--------+---------------+---------------+
|  0x84  | Size          | Content       |
+--------+---------------+---------------+
    1        1 or 2         0 .. 32766
```

A 0x84 field carries its own size, using the encoding of section 2.2.1.3. Within
the static section it may appear before or after other static fields, wherever
its descriptor appeared.

A 0x84 field may also appear **after** the end-of-static tag, as a post-static
field, and a descriptor beyond the fields a method defines is answered there
(section 2.2.7.3). Post-static fields carry the same tag and the same size
encoding as static ones; only their position differs. A reply that carries both
post-static fields and a dynamic section places the dynamic section last.

##### 2.2.8.4 Dynamic Section

```
+--------+----------------------------+
| Tag    | Content                    |
+--------+----------------------------+
    1       to end of host block
```

A dynamic tag is followed by no size. Its content is every remaining byte of the
host block, and the sender therefore MUST place it last. The tag determines what
the content does:

| Tag | Content | Effect |
|---|---|---|
| 0x85 | appended to the message being assembled | The message stays open. The call stays outstanding. |
| 0x88 | appended to the message being assembled | The message is complete and delivered to the caller. The call stays outstanding. |
| 0x86 | appended to the message being assembled | The call is complete. |

The three tags drive two different completion paths in the caller, and they are
not interchangeable. A caller is one of two kinds, fixed by the call it made:

| Tag | Caller waiting for a single result | Caller reading a sequence |
|---|---|---|
| 0x85 | not released; content accumulates | nothing delivered; content accumulates |
| 0x88 | **not released — blocks until the pipe closes** | one message delivered; more may follow |
| 0x86 | released with the accumulated message | **nothing delivered; the sequence ends empty** |

The two bold cells are the failure modes. Both are silent: neither peer signals
anything, and the caller simply never wakes.

A sequence therefore ends with **two** host blocks on the same request
identifier: the last content-bearing block ending in 0x88, then a block whose
body is `87 86` and which carries no content. The first closes the last message
and makes it readable; the second ends the sequence. Without the second block
the caller never learns that the sequence has ended. Section 5.4 shows the pair.

A sender MUST emit these tags as the exact byte values given here. A receiver
MAY dispatch on them by masking — a conforming implementation recognizes
completion as `(tag & 0x8F) == 0x86` and sequence end as `(tag & 0x8F) == 0x88`
— so a sender that sets any other bit of the tag gets behavior this
specification does not define.

Only the first host block of a sequence carries a static section (section
2.2.8.1); every later block begins at the end-of-static tag.

One dynamic message MUST NOT exceed 16,384 bytes of content (section 6.2). A
sender with more to deliver MUST split it into several host blocks, each ending
its own message with 0x88. Consecutive 0x85 blocks accumulate into one message
and overrun the receiver's message buffer at that limit. A receiver that meets a
dynamic message exceeding the bound discards the message and fails the call.

##### 2.2.8.5 Error Field

```
+--------+---------------+
|  0x8F  | Error code    |
+--------+---------------+
    1            4
```

A 0x8F field replaces the reply the call would have returned. Its four
little-endian bytes are:

| Value | Meaning | Origin |
|---|---|---|
| 0xE0000001 | Parameter type does not match the request | client-local |
| 0xE0000002 | Message is not a valid host block | client-local |
| 0xE0000003 | Parameter could not be added, or the send failed | local, either peer |
| 0xE0000004 | Method is not registered on this interface | wire, server-emitted |
| 0xE0000005 | Memory allocation failed | local, either peer |
| 0xE0000006 | Internal error | wire, server-emitted |
| 0xE0000007 | Invalid parameter | wire, server-emitted |
| 0xE0000008 | Send error | local, either peer |
| 0xE0000009 | Chunked field could not be received | wire, server-emitted |
| 0xE000000A | The service refused the call | wire, server-emitted |

The Origin column says where a value comes from. **Wire** values are the ones a
server puts in a 0x8F field; a client MUST accept all five, and a server MUST
NOT emit any other. **Client-local** and **local** values are raised by an
implementation to its own application when its own API call fails, never
transmitted, and listed here only because both peers share one error space.

**The byte shape of an error reply.** The 0x8F field is a static field and the
static section still ends with the end-of-static tag. The whole body of an error
reply is:

```
+--------+---------------+--------+
|  0x8F  | Error code    |  0x87  |
+--------+---------------+--------+
    1            4            1
```

```
8f xx xx xx xx 87
```

and nothing else: no other static field, no post-static field, and no dynamic
section, whatever the call's receive descriptors declared. A client that parsed
the static section looking for the declared fields and never found the
end-of-static tag would block on a call that has already failed. Section 5.7
shows the reply on both bindings.

#### 2.2.9 Iterator Cancel

```
+--------+--------+------------------+--------+
| Class  | Method | RequestId (VLI)  |  0x0F  |
+--------+--------+------------------+--------+
    1        1        1, 2, or 4          1
```

The answer:

```
+--------+--------+------------------+--------+--------+
| Class  | Method | RequestId (VLI)  |  0x87  |  0x88  |
+--------+--------+------------------+--------+--------+
    1        1        1, 2, or 4          1        1
```

A call block whose body is the single byte 0x0F cancels the sequence outstanding
on its Class, Method, and RequestId. It is sent by the caller when it abandons a
sequence before the sequence ends.

The peer MUST answer with a call block repeating that Class, Method, and
RequestId, whose body is `87 88`. The answer is unconditional: it is sent even
when the Class, Method, and RequestId name no sequence the peer is running.
Answering costs one two-byte body, and the caller discards an answer it is not
waiting for (section 3.2.5.2); staying silent leaves a caller that *is* waiting
blocked until the pipe closes. A peer therefore answers first and looks up the
sequence afterwards.

On answering, the peer discards the sequence, if it had one, and MUST send no
further block on that Class, Method, and RequestId. This `87 88` body completes
the call at the caller, unlike the general rule for 0x88 in section 2.2.8.4, and
the request identifier becomes free for reuse.

No other body encoding produces a single 0x0F byte: every send parameter tag
carries data after it, and every receive descriptor has bit 7 set.

## 3 Protocol Details

### 3.1 Common Details

Both peers implement the pipe layer identically. This section specifies that
behavior once; sections 3.2 and 3.3 add the call-layer behavior specific to each
role. Nothing here depends on which binding carries the messages.

#### 3.1.1 Abstract Data Model

Per connection:

- **Binding**: the transport binding in use and its state (section 4).
- **TransportParameters**: PacketSize, MaxBytes, WindowSize, AckBehind,
  AckTimeout, and KeepAlive, as negotiated in section 2.2.3.1.2 and applied by
  the binding.
- **OpenPipes**: the pipe indexes in use, and for each one whether it is opening
  or open. What a pipe is bound to is in section 3.2.1 or 3.3.1.

#### 3.1.2 Timers

None. Every timer in this protocol belongs to a binding (section 4.2.4). A
client MAY apply its own call timeout; the protocol defines none.

#### 3.1.3 Connection Bring-Up

Bring-up begins once the binding reports the link up, and consists of four
connection-level pipe messages, none of them on a pipe. Section 5.1 shows all
four on both bindings:

1. The client sends a type 4 control frame (section 2.2.3.1.3) as soon as the
   binding's framing is up, without waiting to be prompted.
2. The server sends a type 3 control frame (section 2.2.3.1.2). It MUST be the
   server's first pipe message, and the server MUST send it unprompted, without
   waiting for the client to speak first, once the binding's link initialization
   (section 4.2.5 or 4.3.3) has completed. On Select that initialization is
   itself a client-first exchange, and the parameters are the server's first
   output after it.
3. The client applies the parameters and sends a type 1 control frame (section
   2.2.3.1.1).
4. The server echoes the type 1 body exactly.

The connection is ready **at the client** when the echo returns with the body
unchanged.

Steps 1 and 2 are not ordered against each other and may cross on the link. A
server MUST NOT wait for the type 4 before sending the parameters, and MUST
accept one that arrives before it has sent them. Because the client sends its
type 4 before any data packet has reached it, on Select that frame carries an
acknowledgment of 0.

**What the server does with the type 4 frame.** Nothing. It is an announcement,
not a request: the server records nothing, changes no state, and sends no
answer. A second type 4, or none at all, is equally without effect — a server
MUST NOT wait for one, MUST NOT refuse to serve a connection that never sends
one, and MUST NOT treat a repeat as an error.

**The server has no readiness gate.** Bring-up is a client-side sequence. A
server serves every pipe message it receives from the moment the binding reports
the link up, in the order it arrives, whether or not it has yet echoed the
type 1 frame and whether or not the type 1 frame ever arrives. A pipe-open
request or a call that a client pipelines behind its connection request is
answered normally. A server that queued or dropped such messages would hang
against a client that pipelines, and nothing in this protocol would say so.

#### 3.1.4 Higher-Layer Triggered Events

**Sending a pipe message.** The pipe layer hands the binding one pipe unit: the
destination pipe index and the message. What the binding does with it is section
4.

**Closing a pipe.** The peer sends the pipe-close message of section 2.2.4 and
discards the pipe's state. Nothing is said to the binding: a close is a pipe
message on a routing value the binding may not read (section 4.1), and the
binding's reassembly state is keyed on reassembly indexes, which the pipe layer
cannot name (section 2.1.1). A peer that closes a pipe mid-message leaves that
message's context in place; a context is released when its message completes, or
when the connection ends.

#### 3.1.5 Message Processing Events and Sequencing Rules

##### 3.1.5.1 Routing a Pipe Message

The binding delivers a pipe unit. Read the two-byte routing value at the start
of the message and dispatch per section 2.2.2:

| Message | Action |
|---|---|
| Shorter than two bytes | Discard. |
| Routing 0xFFFF | Control frame (section 2.2.3.1). |
| Routing 0x0000 | Pipe-open request (section 3.3.5.1 at a server; discarded at a client). |
| Routing 0x0001 - 0x000F, pipe not open | Discard. No response, and no state changes. |
| Routing 0x0001 - 0x000F, pipe opening | Pipe-open response (section 2.2.3.3). |
| Routing 0x0001 - 0x000F, pipe open, content one byte 0x01 | Pipe close (section 2.2.4). |
| Routing 0x0001 - 0x000F, pipe open, content one byte other than 0x01 | Discard. |
| Routing 0x0001 - 0x000F, pipe open, content empty | Discard: a host block needs at least a Class byte. |
| Routing 0x0001 - 0x000F, pipe open, content two bytes or more | Pipe data (section 2.2.3.4). |
| Routing 0x0010 - 0xFFFE | Discard. Unassigned (section 2.2.2). |

Every discard above is silent. No error message exists at this layer and none is
sent.

Whatever framing field the binding reports, a receiver MUST route on the routing
value. This is the one place that rule is stated; sections 2.2.2, 4.2.2 and
4.3.1 refer to it. Neither binding's field carries a pipe number, and a receiver
MUST NOT read one out of either.

On Select the field is a reassembly index (section 4.2.3): transmit state, taken
free when a message begins and released when it completes. The number says only
which of the sixteen were free at the time, so two messages on one pipe
routinely carry different indexes and two on different pipes carry the same
one.

On Straight the field echoes the pipe message's command byte (section 4.3.1)
and is zero for every message that has none. A receiver that read it as a pipe
number would address every call and every reply to a number that cannot name a
pipe, and apply every close to pipe 1 (section 5.8).

#### 3.1.6 Timer Events

None. See section 4.2.10.

#### 3.1.7 Other Local Events

**Loss of the link.** The binding reports the connection failed. All pipes are
closed, all outstanding calls fail, every open stream is discarded, the binding
is directed to discard any partial message it holds (section 2.1.1), and the
connection is discarded.

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

None beyond section 3.1.2.

#### 3.2.3 Initialization

The client drives its binding's link initialization (section 4.2.5 or 4.3.3),
then performs its half of section 3.1.3.

#### 3.2.4 Higher-Layer Triggered Events

**Opening a service.** The client allocates an unused pipe index, sends a
pipe-open request (section 2.2.3.2), and waits for the pipe-open response and
then the interface table. It MUST NOT issue a call on the pipe before it has
resolved the GUID it needs to an identifier from that table.

**Issuing a call.** The client:

1. Allocates a request identifier that no outstanding call on the pipe is using.
2. Builds the body: send parameters in the order the method defines, then one
   receive descriptor per field it expects back.
3. Replaces any variable-length argument that does not fit in the call block
   with a chunked reference (section 2.2.7.2), allocating a stream identifier
   from NextStreamId.
4. Sends the call block on the pipe, then the stream frames for each chunked
   field, in order, 0xE7 on the last frame of each stream.
5. Records the call in OutstandingCalls with the Class and Method it sent.

Stream frames are one-way and unacknowledged. The client MUST NOT expect any
response to one, and MUST NOT pause between them: it emits the call block and
then every frame of every stream that call referenced, back to back. The reply
to the call arrives after the last frame, because the server holds it until the
field is complete (section 3.3.5.3). A client that sends a large argument as a
series of calls therefore paces itself naturally, each call's reply confirming
that the previous field landed before the next one starts.

**Cancelling a sequence.** The client sends the iterator-cancel body (section
2.2.9) on the Class, Method, and RequestId of the sequence, and treats the call
as complete when the `87 88` acknowledgment arrives.

**Closing a service.** The client sends the pipe-close message (section 2.2.4)
and fails any calls still outstanding on that pipe.

#### 3.2.5 Message Processing Events and Sequencing Rules

##### 3.2.5.1 Receiving an Interface Table

The client stores the GUID-to-identifier assignments for the pipe. An identifier
is meaningful only on that pipe.

If a GUID the client requires is absent from the table, no interface can be
resolved: the client closes the pipe as soon as the table arrives, without
issuing a call. A table MUST therefore carry every GUID its service is reached
through.

##### 3.2.5.2 Receiving a Reply

A reply is matched in three steps:

1. Class against the interface handler registered for the pipe.
2. RequestId against OutstandingCalls.
3. Method against the Method recorded with that outstanding call.

A reply that fails any step is discarded. No error is returned and no indication
reaches the application, and the call it was meant to answer stays outstanding
until the pipe closes.

On a match, the client parses the body per section 2.2.8:

- Static fields are delivered to the caller in order.
- A reply that ends at the end-of-static tag completes the call, which is
  removed from OutstandingCalls (section 2.2.8.2).
- Otherwise the dynamic tag determines completion, per the table in section
  2.2.8.4. On 0x86 the call completes and is removed from OutstandingCalls. On
  0x88 the current message is delivered and the call remains outstanding. On
  0x85 the content is accumulated and the call remains outstanding.
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

**Server obligations defined elsewhere.** This section covers what a server does
with each message it receives. These rules bind a server too and are stated in
the section that owns the field:

| Rule | Section |
|---|---|
| Echo the type 1 control frame body exactly, without parsing it | 2.2.3.1.1 |
| Send the transport parameters as the first pipe message, and what the values mean | 2.2.3.1.2, 3.1.3 |
| Route only on the routing value, never on the binding's framing field | 3.1.5.1 |
| Repeat Class, Method, and RequestId in every reply, or the client discards it | 2.2.5.1, 3.2.5.2 |
| End every reply's static section with the end-of-static tag | 2.2.8.2 |
| Emit only the first block of a sequence with a static section, and end a sequence with two blocks | 2.2.8.1, 2.2.8.4 |
| Emit only the five wire error values, in the shape `8f xx xx xx xx 87` | 2.2.8.5 |
| Answer an iterator cancel unconditionally | 2.2.9 |
| Enforce every protocol and binding bound | 6.2 |
| Never interleave the frames of two pipe units | 2.1.1 |

#### 3.3.1 Abstract Data Model

In addition to section 3.1.1:

- **PipeHandlers**: for each open pipe, the service bound to it. One service may
  be bound to more than one pipe.
- **SessionState**: per connection, the identity established by the service that
  authenticates it. Every pipe on the connection shares it. It is set after the
  pipe carrying the authentication is already open, so a service MUST NOT bind
  identity at pipe-open time.
- **OpenStreams**: per connection, the chunked fields being received, keyed by
  stream identifier, each recording the pipe and the outstanding call that
  referenced it. The namespace is per connection (section 2.2.7.2), so one
  stream identifier names one stream however many pipes are open.
- **OutstandingCalls**: for each open pipe, the calls received and not yet
  answered, keyed by request identifier.

#### 3.3.2 Timers

None beyond section 3.1.2.

#### 3.3.3 Initialization

The server accepts the link, completes its binding's link initialization
(section 4.2.5 or 4.3.3), and sends the transport parameters as its first pipe
message. It then answers control frames per section 2.2.3.1.

The server serves pipe messages from that moment on and holds no readiness gate
of its own (section 3.1.3).

#### 3.3.4 Higher-Layer Triggered Events

**Pushing a message.** A service may send a call block on a pipe at any time,
repeating the Class, Method, and RequestId of a call the client left outstanding
(section 3.2.5.3). Only the first block of such a sequence carries a static
section (section 2.2.8.1), and sends are serialized per section 2.1.1.

#### 3.3.5 Message Processing Events and Sequencing Rules

##### 3.3.5.1 Receiving a Pipe Open Request

The server:

1. Validates the request. A request naming index 0, an index above 15, or
   an index that is already **opening or open** is discarded: no response, no
   table, and no change to the state of the pipe that is already there.
   Rebinding a live pipe would strand every call outstanding on it, and index 0
   can never be addressed, its routing value being the pipe-open request itself.
   A malformed request (section 2.2.3.2) is discarded on the same terms. An
   index that was opened and has since been closed is not in use and may be
   opened again.
2. Sends the pipe-open response (section 2.2.3.3).
3. Binds the named service to the pipe index. Matching on the service name is
   case-insensitive. The same service may be bound to several pipes on one
   connection; each binding is independent and each pipe gets its own table.
4. Sends the interface table for that service (section 2.2.6), unless step 5
   withholds it.
5. Withholds the table when the server will not serve the request:

| Condition | Response | Table |
|---|---|---|
| Service known, version served, ready | sent | sent |
| Service known, version served, not otherwise ready | sent | **sent** |
| Service known, version **not** served | sent | withheld |
| Service name unknown | sent | withheld |

There is no encoding for refusing an open (section 2.2.3.3), so the withheld
table is the refusal. The pipe opens either way. A client that receives no table
resolves no interface, blocks until its own timeout expires, and abandons the
pipe (section 3.2.5.1); the protocol defines no such timeout.

A service that is merely not yet ready is a different case and MUST still get
its table, because the client needs the identifiers before it can call anything
at all, and readiness is a condition the service resolves on its own.

Section 5.2 shows a request, its response, and a table.

##### 3.3.5.2 Receiving a Call Block

1. If Class is 0xE0 or above (section 2.2.5), process the block as a stream
   frame (section 3.3.5.3). The test is on the whole byte, not on any bit of it:
   `Class >= 0xE0`. A server that tested `(Class & 0xE0) != 0` would send every
   call on an interface identifier from 0x20 up into the stream path, where it
   is appended to a stream that does not exist and discarded.
2. If the block's Class is 0x00, discard it. A server never receives an
   interface table.
3. If RequestId names a call already outstanding on the pipe, discard the block
   (section 2.2.5.1).
4. If the body is the single byte 0x0F, answer the iterator cancel per section
   2.2.9.
5. Otherwise resolve Class to an interface and Method to a method, parse the
   body per section 2.2.7, and execute.
6. Build the reply body per section 2.2.8 and send it with the Class, Method,
   and RequestId of the call, unchanged.

Errors are decided in this order, and the first that applies decides the answer:

| Condition | Answer |
|---|---|
| Class resolves to no interface in the table published for this pipe | **none** — discard the call |
| Body parsing stops at an unassigned send-parameter tag (section 1.8) | error 0xE0000007 |
| Body breaches a bound of section 6.2 — over 16 send parameters, over 16 receive descriptors, a variable field longer than the bytes remaining, a chunked reference over the length bound | error 0xE0000007, except the chunked length, which is 0xE0000009 |
| Method is not implemented on the resolved interface | error 0xE0000004 |
| The service declines to run the method | error 0xE000000A |
| The method runs and fails inside the server | error 0xE0000006 |
| The method runs and returns nothing | **none** — no reply is sent |
| The method runs and returns | the reply of section 2.2.8 |

Where an error is answered, the parameters parsed so far are discarded and the
method does not run. The reply body is the error field alone, in the shape
section 2.2.8.5 gives.

A Class that resolves to no interface cannot be answered at all: the client
matches a reply on Class before anything else (section 3.2.5.2), so every reply
the server could send would be discarded. Silence and an error are equally
invisible there, and the call stays outstanding at the client until the pipe
closes.

A reply of a shape the receive descriptors did not ask for fails the call at the
client with a parameter type mismatch.

A server MUST bound the number of calls it will hold outstanding on a connection
(section 6.2). RequestId is a VLI reaching 1,073,741,822, and every call
accepted allocates state until it is answered. Calls beyond the bound are
discarded.

##### 3.3.5.3 Receiving a Stream Frame

The server appends the field bytes to the stream named by byte 1, and marks the
stream complete on class 0xE7. It MUST NOT reply.

A frame is accepted only when byte 1 names a stream in OpenStreams **and** the
frame arrived on the pipe whose call opened that stream. Any other frame is
discarded: one naming an identifier no call referenced, one naming a live stream
but arriving on another pipe, and one naming a stream that has already seen its
0xE7 frame. A stream identifier stays bound to its call, and unusable by any
other, until the stream is complete or discarded — which may be after the call
has been answered.

**When the call is answered.** Stream frames arrive after the call block that
references them. A method whose result depends on a chunked field MUST NOT
answer the call block until every stream that call referenced has seen its 0xE7
frame. The reply is built and sent then, and carries one field per receive
descriptor the call declared, exactly as any other reply does.

Deferring the reply is what paces the sender. The caller blocks on the reply, so
holding it until the field is complete stops the next call's field from
overlapping the one still arriving; answering early lets two streams for the
same argument interleave in the sender's buffer. It is also what makes a failure
reportable — a reply not yet sent can carry an error field, and one already sent
cannot.

A stream that completes short of its declared Length fails the operation waiting
on it. The server discards the stream, discards the operation it fed, and
answers the call that referenced it with the error field 0xE0000009. A call that
referenced several streams is answered once, on the first failure.

A stream the pipe close ends before it completes is discarded with the rest of
the pipe's state and answers nothing: the client is not listening on a pipe it
closed (section 3.3.5.4). A connection failure ends everything the same way
(section 3.1.7).

A method whose result does **not** depend on the chunked field — one that only
files the bytes away — MAY be answered as soon as the call block arrives. The
call is then complete, its request identifier is free, and a later stream
failure has nowhere to be reported: the server discards the stream silently. A
service that needs failures reported MUST NOT answer early.

##### 3.3.5.4 Receiving a Pipe Close

The server releases the service bound to the pipe, discards the pipe's state,
discards any streams open on it, and fails any calls outstanding on it. The
binding keeps whatever it holds: it cannot be told to discard by pipe (section
2.1.1). No response is sent.

When a close leaves no service pipe open on a connection that had at least one,
the server closes the connection. A connection on which no service pipe has yet
been opened is held.

**Closing the connection** means releasing the underlying link, and each binding
releases it the same way whatever caused the release:

| Binding | Action |
|---|---|
| Straight | Close the TCP socket. |
| Select | Close the underlying link — the TCP socket of a bridged endpoint, or drop the carrier on a physical serial port. No packet, control frame, or pipe message announces it, and none is defined. |

The peer learns of it as the link-failure event of section 4.1. On a physical
serial link that a server cannot drop, the connection is released locally and
the peer learns nothing until its own timers expire; a deployment that needs a
positive signal must obtain it below this protocol.

#### 3.3.6 Timer Events

None beyond section 3.1.6.

#### 3.3.7 Other Local Events

None.

## 4 Transport Bindings

### 4.1 Common Requirements

A binding implements the service of section 2.1.1 and nothing above it. It MUST
NOT read the routing value, MUST NOT interpret the content of a pipe message,
and MUST deliver the message's bytes exactly as they were submitted.

The two bindings are not negotiated and cannot be mixed on one connection.
Nothing on the wire announces the choice and no message identifies it.

This protocol defines no timing. A peer MAY pace its output — a deployment on a
slow serial link commonly does, to keep a receiver's buffers from filling — and
a peer MUST NOT infer anything from how long a message takes to arrive or from
how closely two messages follow each other. The only timers are the Select
binding's (section 4.2.4), and the only deadline is the framing handshake's
(section 4.2.5.1).

End of the underlying link — a closed socket, a dropped carrier, a read that
returns nothing — is reported to the pipe layer as the failure event of section
2.1.1, whatever else the binding is doing at the time. This is the only event
that releases a connection's state when the peer disappears without closing its
pipes.

Each binding is served on its own endpoint, and the endpoint selects the
binding:

| Binding | Endpoint |
|---|---|
| Straight | TCP port 569. Fixed. |
| Select | A serial port, or a TCP endpoint bridged to one (section 4.2.5.2). Deployment-defined. |

A server that offers both MUST choose the binding from the endpoint that
accepted the connection. Forwarding one endpoint to the other therefore changes
the binding a connection is served with: a redirect from 569 to the Select
endpoint answers a Straight client with packets, which it cannot parse.

### 4.2 Select Binding

The Select binding carries pipe units over one bidirectional byte stream, and
provides its own framing, error detection, retransmission, and fragmentation. It
takes nothing from the link beyond ordered delivery of bytes.

Both peers may transmit at any time after link initialization completes.

**What each peer must implement.** The binding is symmetric on the wire but not
in practice: a peer's obligations divide into what it must do to be understood
and what it must do to recover from loss. Both peers implement the whole of the
first group. The second group is a sender's own affair — a peer that omits it
still interoperates, and loses only its ability to recover.

| Machinery | Section | Required of |
|---|---|---|
| Packet framing, escape encoding, transform, check field | 4.2.1 | both peers |
| Pipe frames, fragmentation, reassembly | 4.2.2, 4.2.8, 4.2.9 | both peers |
| Emitting an acknowledgment for every packet received | 4.2.4, 4.2.7 | both peers |
| Reading and honoring a received acknowledgment | 4.2.7 steps 6-7 | a peer that retransmits |
| Duplicate suppression on ExpectedSequence | 4.2.7 step 8 | a peer whose own peer retransmits |
| UnacknowledgedPackets, retransmission timer, window | 4.2.3, 4.2.6, 4.2.10 | a peer that retransmits |
| Emitting a negative acknowledgment | 4.2.1.5 | no peer; it is optional |
| Handling a received negative acknowledgment | 4.2.1.5, 4.2.7 step 5 | a peer that retransmits |

A peer that does not retransmit MUST still acknowledge everything it receives,
because its peer may retransmit, and MUST still reverse the transform on the
bytes of every packet it reads. It simply never fills UnacknowledgedPackets, and
the loss of one of its packets fails the connection rather than being repaired.

The stream carries packets (section 4.2.1). Every packet ends with the byte
0x0D, which never occurs anywhere else in a packet. A receiver frames the stream
by scanning for 0x0D, and MUST treat everything since the previous terminator as
one packet.

Both peers must be able to transmit and receive all 256 byte values except that
this binding reserves 0x0D as a terminator and escapes it inside packets;
implementations therefore interoperate over links that are transparent to eight
bits.

#### 4.2.1 Packet

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
| Payload | 0 - N | Pipe frames (section 4.2.2), escape-encoded per section 4.2.1.1. |
| Check | 4 | Check field over bytes 0 through n-6, masked per section 4.2.1.3. |
| Terminator | 1 | 0x0D. |

The Seq byte selects the packet type:

| Value | Type |
|---|---|
| 0x41 | Acknowledgment packet (section 4.2.1.4) |
| 0x42 | Negative acknowledgment packet (section 4.2.1.5) |
| any other | Data packet |

In a data packet, Seq has bit 7 set and bits 6-0 hold the sequence number, after
the reserved-value transform of section 4.2.1.2 is reversed. Sequence numbers
run 0 through 127 and wrap.

The minimum packet is seven bytes: Seq, Ack, an empty payload, the check field,
and the terminator. The maximum is the negotiated PacketSize (section
2.2.3.1.2), which counts every byte of the packet including the terminator.

An empty packet — a terminator with nothing before it — is used during link
initialization (section 4.2.5) and carries no other meaning.

##### 4.2.1.1 Escape Encoding

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
section 4.2.1.2, and the check field uses the mask in section 4.2.1.3.

The encoding expands the payload, at worst to twice its length. A sender that
fragments a message MUST size each fragment against the encoded length, because
expansion depends on content (section 4.2.8).

##### 4.2.1.2 Reserved-Value Transform

The Seq and Ack bytes, the PipeHeader byte, and — when HasLength is set — the
frame length byte (section 4.2.2) are not escape-encoded. They avoid the
reserved values of section 4.2.1.1 by a reversible transform instead: if the
byte would be a reserved value, it is transmitted exclusive-ORed with 0xC0.

Every byte in these positions has bit 7 set, so only three reserved values can
occur there:

| Value | Transmitted as |
|---|---|
| 0x8D | 0x4D |
| 0x90 | 0x50 |
| 0x8B | 0x4B |

A receiver MUST reverse the transform: a received 0x4D, 0x50, or 0x4B in one of
these positions is exclusive-ORed with 0xC0 to recover the value. All other
values pass through unchanged in both directions.

The transform applies to the byte after bit 7 has been set, so a sender computes
the Seq byte as `transform(sequence | 0x80)` and the Ack byte as
`transform(acknowledgment | 0x80)`.

A byte in these positions is never escape-encoded. A receiver reverses the
transform on them and unescapes the rest of the payload.

##### 4.2.1.3 Check Field

The check field is a 32-bit cyclic redundancy check computed over the packet
bytes as they appear on the wire — the Seq byte, the Ack byte, and the encoded
payload — and transmitted little-endian, then masked.

The register is initialized to 0, no final inversion is applied, and the
polynomial is 0x248EF9BE with a bitwise complement on the odd path:

```
register = 0
for each byte b of the input:
    register = register XOR b
    repeat 8 times:
        if (register & 1) != 0:
            register = ~((register ^ 0x248EF9BE) >> 1)
        else:
            register = register >> 1
```

The equivalent table-driven form, where each table entry is the inner
eight-iteration loop run with the register seeded to the entry's index and no
input byte mixed in:

```
table[i] = eight iterations of the inner loop, starting from register = i
crc      = 0
for each byte b of the input:
    crc = (crc >> 8) ^ table[(b ^ (crc & 0xFF)) & 0xFF]
```

All arithmetic is on 32-bit unsigned values.

This is bit-for-bit the standard reflected CRC-32 — polynomial 0xEDB88320,
register initialized to 0, no final inversion. The complement on the odd path
and the polynomial 0x248EF9BE above cancel exactly, for every input. An
implementation MAY use a stock CRC-32 routine configured that way instead:

```
register = 0
for each byte b of the input:
    register = register XOR b
    repeat 8 times:
        if (register & 1) != 0:
            register = (register >> 1) XOR 0xEDB88320
        else:
            register = register >> 1
```

Note that this is not the CRC-32 of Ethernet or zip, which initialize the
register to 0xFFFFFFFF and invert the result.

The four bytes of the little-endian result are then masked: any byte whose value
is 0x1B, 0x0D, 0x10, 0x0B, 0x8D, 0x90, or 0x8B is transmitted OR 0x60. The mask
is not reversible. A receiver MUST compute the check value over the received
bytes, apply the same mask to its own result, and compare the masked forms.

Worked values:

| Packet bytes checked | Masked check field |
|---|---|
| `80 80 e0 03 00 ff ff 04` | `fc 03 18 a0` |
| `80 80 e0 17 00 ff ff 03 00 04 00 00 00 04 00 00 1b 32 00 00 00 01 00 00 00 58 02 00 00` | `38 c9 9a 7e` |

##### 4.2.1.4 Acknowledgment Packet

```
+--------+--------+--------+------+
|  0x41  |  Ack   | Check  | 0x0D |
+--------+--------+--------+------+
    1        1        4        1
```

An acknowledgment packet carries no payload and no sequence number. Its Ack
field is built exactly as a data packet's, `transform(ExpectedSequence | 0x80)`.
It is not itself acknowledged and does not advance the sender's sequence space.

Its Seq byte is the literal 0x41 and takes neither bit 7 nor the transform: no
data packet can collide with it, because a data packet's Seq byte always has
bit 7 set.

This is also the packet a keep-alive timer sends (section 4.2.4); nothing
distinguishes a keep-alive from any other acknowledgment, and none is needed.

An acknowledgment of sequence 1, whose check field is computed over the two
bytes `41 81`:

```
41 81 f2 cd dd 73 0d
```

##### 4.2.1.5 Negative Acknowledgment Packet

```
+--------+--------+--------+------+
|  0x42  | Seq    | Check  | 0x0D |
+--------+--------+--------+------+
    1        1        4        1
```

Byte 1 is a sequence number, not an acknowledgment, and it is built and parsed
exactly as the Ack byte of any other packet: the sender emits
`transform(sequence | 0x80)`, and a receiver reverses the transform of section
4.2.1.2 and then masks with 0x7F. The recovered value is the last sequence
number the sender received intact; section 4.2.7 step 5 says what the peer does
with it.

Byte 0 is the literal 0x42 and takes neither bit 7 nor the transform.

A negative acknowledgment naming sequence 13. The value `0x8d` is a reserved
value and is transmitted exclusive-ORed with 0xC0:

```
42 4d aa 70 22 ca 0d
```

A receiver that discards a damaged packet silently, as section 4.2.7 step 3
requires, never emits this packet. It arises only from a peer that tracks
sequence gaps and asks for the missing range. Every peer that retransmits MUST
handle one on receipt (section 4.2).

#### 4.2.2 Pipe Frame

A packet payload holds one or more pipe frames laid end to end. A frame carries
one pipe unit, or part of one.

```
 0            1              1 or 2
+------------+--------------+------------------+
| PipeHeader | Length (opt) | Content          |
+------------+--------------+------------------+
      1          0 or 1           variable
```

PipeHeader, after the reserved-value transform of section 4.2.1.2 is reversed:

| Bits | Mask | Field | Description |
|---|---|---|---|
| 3-0 | 0x0F | ReassemblyIndex | The reassembly context this message occupies (sections 3.1.5.1, 4.2.3). Not a pipe index. |
| 4 | 0x10 | HasLength | A length byte follows the header and gives the frame's content length. |
| 5 | 0x20 | Continuation | No length byte follows. The frame's content runs to the end of the packet, or until BytesOutstanding for the message reaches zero, whichever comes first (section 4.2.9 step 3). |
| 6 | 0x40 | LastData | This frame ends the pipe message. |
| 7 | 0x80 | — | Always set. |

Exactly one of HasLength and Continuation MUST be set. A receiver that meets a
header with both set MUST parse the frame as Continuation; with neither set it
MUST parse it as HasLength. Guessing differently costs framing for the rest of
the packet, because the next frame begins at the first byte this one did not
consume (section 4.2.9).

When HasLength is set, the byte after the header — also subject to the transform
of section 4.2.1.2 — has bit 7 set, and bits 6-0 give the number of content
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

ContentLength is frame content, not a frame header field: it is escape-encoded
per section 4.2.1.1 like the message bytes that follow it. A message of 11
bytes therefore puts `1b 33` on the wire, not `0b`.

A pipe message is complete when ContentLength bytes have arrived, and a receiver
MUST use that count — not the LastData flag — to decide where one message ends
and the next begins. LastData marks the final frame of a transmission sequence
and can be clear on the frame that completes a message. A receiver that ends a
message on LastData joins two messages into one buffer with a frame header
embedded in it.

A sender MUST set Continuation on every frame it emits, on every message,
including short ones that would fit a HasLength frame. This is not a preference:
a peer built against the Continuation form may misroute a message that arrives
in the HasLength form, and the failure shows up on one message rather than all
of them. A receiver MUST accept HasLength frames all the same, because it cannot
know how its peer was built.

The binding therefore never inspects a message to choose a frame form, which it
could not do in any case: it holds no pipe state, and by section 2.2.2 the
message alone does not say what it is.

#### 4.2.3 Abstract Data Model

Per connection:

- **SendSequence**: the sequence number for the next data packet. Initialized to
  0 and incremented modulo 128 for every data packet sent.
- **ExpectedSequence**: the sequence number the peer is expected to send next,
  transmitted in the Ack field. Initialized to 0.
- **UnacknowledgedPackets**: data packets sent and not yet acknowledged, in
  order.
- **EffectiveWindow**: the number of packets the sender will currently hold
  outstanding. Initialized to WindowSize.
- **Reassembly**: sixteen contexts, one per reassembly index, each idle or
  holding a partially received message as:
  - **Buffer**: message bytes accumulated so far.
  - **BytesOutstanding**: message bytes still expected. Non-zero exactly while
    the context is active.

There are **sixteen** reassembly contexts per connection, one per reassembly
index, and none per pipe. The index in a frame header is the only thing that
separates two partially received messages (section 2.1.1), so a receiver keys
reassembly on it and on nothing else. It never routes on it: the completed
message's routing value is the destination (section 3.1.5.1).

A context is released the moment its message completes. A frame that opens a new
message on a context whose BytesOutstanding is non-zero means the peer reused an
occupied index. The receiver MUST discard that context's Buffer and the frame
that arrived, and set its BytesOutstanding to zero; both messages are lost
either way, and keeping the buffer only spreads the damage into a third.

The binding reads TransportParameters (section 3.1.1): PacketSize and MaxBytes
bound what it emits, WindowSize and AckBehind bound what it holds outstanding
and when it acknowledges, and AckTimeout and KeepAlive drive its timers.

#### 4.2.4 Timers

- **Retransmission timer.** Runs while UnacknowledgedPackets is non-empty, with
  a period of AckTimeout.
- **Acknowledgment timer.** Optional. A receiver that does not acknowledge every
  packet immediately MUST acknowledge within AckTimeout, and MUST acknowledge
  after AckBehind unacknowledged packets have accumulated.
- **Keep-alive timer.** Present only when KeepAlive was negotiated. On expiry
  the peer sends an acknowledgment packet.

#### 4.2.5 Link Initialization

The binding needs a byte stream that delivers bytes in order and is transparent
to eight bits. Section 4.2.5.1 defines the exchange that starts the framing on
it. Section 4.2.5.2 defines what a server does when that stream is reached
through a telnet-bridged endpoint instead of a serial port.

##### 4.2.5.1 Framing Handshake

The handshake runs before any packet is exchanged.

1. The client sends an empty packet — the single byte 0x0D.

```
+------+
| 0x0D |
+------+
    1
```

2. The server responds with the three ASCII characters `COM` followed by 0x0D.
   This response selects binary packet framing on the link.

```
+------+------+------+------+
| 'C'  | 'O'  | 'M'  | 0x0D |
+------+------+------+------+
    1      1      1      1
```

The client MUST abort the connection if the response in step 2 does not arrive
within 20 seconds, and MUST abort if it is anything other than `COM` followed by
0x0D.

The server's side of the same exchange:

- It waits for the client's empty packet and MUST NOT write anything before it,
  not even the transport parameters. The `COM` reply is the server's first
  output on the link, and section 3.1.3's rule that the server speaks unprompted
  applies to the pipe layer, which is above this exchange.
- Bytes received before the empty packet are discarded, except telnet
  negotiation on a bridged endpoint, which is handled per section 4.2.5.2. A
  first byte that is not 0x0D is not an error: the server keeps scanning for
  0x0D and treats everything before it as noise from the modem or the bridge,
  which is where such bytes come from.
- The server MUST bound how long it waits and how many bytes it discards, and
  MUST release a link that exceeds either. A half-open serial or bridged
  connection is otherwise retained forever. The bound is deployment-defined; the
  protocol defines no timeout here, and the client's 20 seconds is the client's
  own.

After step 2 the link is up in both directions and bring-up proceeds per section
3.1.3. The server's transport parameters travel in a data packet with sequence
number 0. The client's type 4 frame is sent on receipt of the `COM` reply, so it
is its sequence 0 and carries an acknowledgment of 0, and it may cross the
parameters on the link.

##### 4.2.5.2 Telnet-Bridged Links

A Select link is commonly reached over TCP rather than over a physical serial
port, with a modem emulator bridging the two. The client's modem command
exchange — dialling, `CONNECT` — is between the client and that emulator and
never reaches the server. What reaches the server is the byte stream the emulator
opens. A bridge configured for raw passthrough opens it with the framing
handshake of section 4.2.5.1 and nothing else; one configured for telnet opens it
with option negotiation first.

That negotiation is not part of MOS RPC. A server MUST remove it from the stream
before framing, and MUST refuse every option:

| Received | Answer | Stream |
|---|---|---|
| `IAC DO x` | `IAC WONT x` | The three bytes are removed. |
| `IAC WILL x` | `IAC DONT x` | The three bytes are removed. |

`IAC` is 0xFF, `WILL` 0xFB, `WONT` 0xFC, `DO` 0xFD, `DONT` 0xFE. Every other
`IAC` sequence is removed without an answer, and a receiver MUST know how far to
remove:

| Sequence | Removed | Yields |
|---|---|---|
| `IAC` + WILL, WONT, DO or DONT + option | 3 bytes | nothing |
| `IAC SB` … `IAC SE` | everything from `IAC SB` through the terminating `IAC SE` | nothing |
| `IAC IAC` | 2 bytes | one 0xFF data byte |
| `IAC` + any other command | 2 bytes | nothing |

`SB` is 0xFA and `SE` is 0xF0.

A worked removal. The bridge opens the stream with these bytes:

```
ff fd 03 ff fb 01 ff fa 18 00 41 ff f0 ff f1 0d
```

| Bytes | Handling | Answer |
|---|---|---|
| `ff fd 03` | `IAC DO 3` — 3 bytes removed | `ff fc 03` (`IAC WONT 3`) |
| `ff fb 01` | `IAC WILL 1` — 3 bytes removed | `ff fe 01` (`IAC DONT 1`) |
| `ff fa 18 00 41 ff f0` | `IAC SB` … `IAC SE` — all 7 bytes removed | none |
| `ff f1` | `IAC` + another command — 2 bytes removed | none |
| `0d` | not telnet | — |

What reaches the framing scanner is the single byte `0d`: the client's empty
packet, and step 1 of section 4.2.5.1. The server answers `COM` 0x0D and the
handshake is complete.

Removing the wrong number of bytes leaves negotiation data in the stream, where
the framing scanner reads it as packet content. Removing six bytes instead of
seven for the subnegotiation above would leave `f0` there, and the scanner would
read it as the first byte of a packet that never verifies; a stray 0x0D left
behind ends a short packet that fails its check field, and one consumed by
mistake makes the scanner swallow the client's first real packet.

Negotiation is confined to the start of the connection, before the framing
handshake of section 4.2.5.1 completes. Once the transport is running a server
MUST apply no telnet processing at all: the data phase is eight-bit transparent,
0xFF carries no telnet meaning, and neither peer doubles or collapses it. Packet
payloads contain 0xFF as a matter of course — the routing value of every control
frame is 0xFFFF — so a receiver that went on interpreting it as `IAC` would
consume two bytes of a control frame and lose framing.

#### 4.2.6 Sending a Packet

To send a data packet:

1. Take the sequence number from SendSequence and increment SendSequence modulo
   128.
2. Build the Seq byte as `transform(sequence | 0x80)` and the Ack byte as
   `transform(ExpectedSequence | 0x80)`, per section 4.2.1.2.
3. Escape-encode the payload per section 4.2.1.1.
4. Compute and mask the check field per section 4.2.1.3.
5. Append the terminator 0x0D.
6. Append the packet to UnacknowledgedPackets and start the retransmission timer
   if it is not running.

A sender MUST NOT have more than EffectiveWindow packets in
UnacknowledgedPackets, and SHOULD NOT hold more than MaxBytes outstanding,
counting the encoded bytes of every packet in that list. A sender MUST NOT emit
a packet longer than PacketSize.

Acknowledgment packets are built the same way but take no sequence number, are
not added to UnacknowledgedPackets, and do not start the retransmission timer.

#### 4.2.7 Receiving a Packet

1. Read bytes until 0x0D. The bytes before it are the packet. The terminator is
   not part of it, so the packet counted here is one byte shorter than the seven
   of section 4.2.1, which draws the terminator inside it.
2. If fewer than six bytes were read before the terminator, discard the packet.
3. Verify the check field per section 4.2.1.3. On mismatch, discard the packet.
   No error is signaled to the peer.
4. Reverse the transform of section 4.2.1.2 on bytes 0 and 1 to recover the
   packet type, the sequence number, and the acknowledgment.
5. For a negative acknowledgment packet, byte 1 is a sequence number and not an
   acknowledgment (section 4.2.1.5): recover it, retransmit every packet in
   UnacknowledgedPackets from the following sequence number onwards, release
   nothing, and stop.
6. Release from UnacknowledgedPackets every packet whose sequence number is
   before the received acknowledgment, and stop the retransmission timer if none
   remain. A sequence number `s` is before an acknowledgment `a` when
   `(a - s) mod 128` lies in 1 through WindowSize. An acknowledgment outside
   that range names a packet already released and MUST be ignored.
7. For an acknowledgment packet, stop here.
8. If the received sequence number is not ExpectedSequence, the packet is a
   duplicate or has arrived out of order. Acknowledge with the current
   ExpectedSequence, discard the packet without decoding its payload, and stop.
   Without this step a retransmission is delivered a second time: at a message
   boundary it repeats a whole pipe message, and mid-message its bytes are
   appended to the open context's Buffer and the message completes early with
   duplicated content.
9. Set ExpectedSequence to the received sequence number plus 1, modulo 128. Do
   this before processing the payload, so that any packet generated by the
   payload carries the updated value. Then acknowledge, per section 4.2.4:
   either in an acknowledgment packet, or by carrying ExpectedSequence in a data
   packet the receiver is about to send anyway. Both are always permitted. A
   peer with traffic of its own piggybacks most of its acknowledgments, and a
   peer that emits a standalone acknowledgment for every packet the moment it
   arrives — before processing the payload, having already advanced
   ExpectedSequence — is equally conforming and costs only the extra packets.
10. Decode the payload per section 4.2.9 and deliver the pipe units it
    completes.

#### 4.2.8 Fragmenting a Pipe Message

A pipe message that does not fit in one packet is split across consecutive
packets. Section 5.5 works one through, with the receiver's state at each step:

```
Frame 1     Continuation set, LastData clear
            PipeHeader | ContentLength | bytes[0..a]

Frame 2..k-1  Continuation set, LastData clear
            PipeHeader | bytes[a..b]

Frame k     Continuation set, LastData set
            PipeHeader | bytes[..end]
```

Rules a sender MUST observe:

- ContentLength on frame 1 is the length of the **entire** message. A length
  covering only frame 1's own bytes completes the receiver's buffer early; the
  receiver then takes no bytes from the following frame and makes no progress.
- Frames after the first carry no length field (section 4.2.2).
- Every frame carries the same reassembly index, taken free when frame 1 is
  built and released when frame k is sent (section 3.1.5.1).
- Each fragment goes in its own packet, with its own sequence number, in order.
- No frame of another message on the same reassembly index may appear between
  them (section 2.1.1).
- Fragment sizes MUST be chosen so that the **encoded** packet fits PacketSize.
  A fixed allowance for expansion is not sufficient: a fragment dense in the
  values of section 4.2.1.1 doubles in length, and a packet over PacketSize is
  discarded by the receiver without notice. A sender picks the largest fragment
  whose fully built packet fits.
- A sender MUST NOT place a fragment boundary between an escape byte and the
  byte it escapes. Where the largest fitting fragment would do so, the sender
  shortens it by one message byte. A packet's encoded payload therefore always
  ends on a complete escape pair, and a receiver never has to hold a half-decoded
  pair across a packet boundary.

A receiver MUST discard a packet whose encoded payload ends with an unpaired
escape byte. No conforming sender produces one, and there is no state in which
the missing byte could arrive: the following packet begins with a pipe frame
header, not with the tail of the previous payload.

#### 4.2.9 Decoding a Payload and Reassembling Messages

Decoding is per packet and keeps no state across packets: an escape pair never
straddles a packet boundary (section 4.2.8). Decode the payload per section
4.2.1.1. A payload ending with an unpaired escape byte is malformed; discard the
packet.

Message reassembly does carry across packets, in the sixteen contexts of section
4.2.3.

For each pipe frame in the decoded payload:

1. Reverse the transform of section 4.2.1.2 on the header byte, and on the
   length byte if HasLength is set, to recover ReassemblyIndex and the flags.
   ReassemblyIndex selects the context every step below works on.
2. Determine the frame's content bytes: the bytes named by the length byte if
   HasLength is set, otherwise the rest of the payload.
3. If the context's BytesOutstanding is zero, this frame opens a message: read
   ContentLength from the first two content bytes, set BytesOutstanding to it,
   and take up to that many of the bytes that follow. Discard the message if
   ContentLength exceeds 65,532 (section 2.2.2). Otherwise the frame continues
   the message already open on that context: take up to BytesOutstanding bytes,
   with no length field present.
4. Append the taken bytes to the context's Buffer and reduce its
   BytesOutstanding by the number taken. A frame carrying more bytes than
   BytesOutstanding is truncated to it (section 6.2); the surplus belongs to no
   message and the receiver MUST NOT read a frame header out of it.
5. If BytesOutstanding is now zero, the message is complete: deliver Buffer as a
   pipe unit and clear the context.
6. Advance to the next frame at the first byte the current frame did not consume.

A receiver MUST NOT assume a continuation frame owns the rest of the packet: a
second frame can follow it in the same payload, and consuming to the end of the
payload discards it.

#### 4.2.10 Timer Events

**Retransmission timer expiry.** The sender retransmits every packet in
UnacknowledgedPackets, in sequence order, and halves EffectiveWindow to a
minimum of one packet.

An acknowledgment that releases at least one packet is progress: it restores
EffectiveWindow to WindowSize and clears the expiry count. After 12 consecutive
expiries without it, the sender terminates the connection: it releases the link
as section 3.3.5.4 describes and reports the failure event of section 2.1.1 to
the pipe layer. Nothing is sent to the peer, which has by then either
disappeared or stopped being reachable.

**Keep-alive timer expiry.** The peer sends an acknowledgment packet.

### 4.3 Straight Binding

The Straight binding carries pipe units over a TCP connection. TCP already
delivers bytes in order, once, and checked, so the binding adds only the
boundaries the service of section 2.1.1 requires: there are no sequence numbers,
no acknowledgments, no escape encoding, no check field, and no terminator.

Both peers may transmit at any time. TCP is full duplex, the binding adds no
turn-taking of its own, and the ladder in section 4.3.3 orders bring-up only.
A server pushes an unsolicited message (section 3.3.4) without waiting to be
spoken to.

#### 4.3.1 Record

```
 0            2        3
+------------+--------+------------------------------+
| TotalLength| Cmd    | Pipe message                 |
+------------+--------+------------------------------+
      2          1               0 .. 65532
```

| Field | Size | Description |
|---|---|---|
| TotalLength | 2 | Little-endian byte count of the whole record, counting these two bytes. |
| Cmd | 1 | A copy of the pipe message's command byte, for the messages the pipe layer builds with one: `0x01` on a pipe close (section 2.2.4). Zero on everything else — every control frame, every pipe open, every host block — and zero from a sender that has no command byte to echo. It duplicates a byte already in the message, so a receiver MUST ignore it and read the command from the content. |
| Pipe message | variable | One complete pipe message (section 2.2.2). |

One record carries one pipe unit, always whole. There is no fragmentation and no
reassembly: a message the Select binding must split across several packets goes
into one record here, and a sender MUST NOT split one message across records. A
receiver takes each record as a complete pipe message, so a second record is
parsed as a fresh routing value and host block.

PacketSize does not bound a record. TotalLength is the only ceiling, so a record
holds at most 65,532 bytes of pipe message.

#### 4.3.2 Abstract Data Model

Per connection:

- **ReceiveBuffer**: bytes received and not yet consumed by a whole record.

Nothing else. Every value in the transport parameters (section 2.2.3.1.2) is
inert on this binding, and the connection keeps no sequence, window, or timer
state.

#### 4.3.3 Link Initialization

None. The client connects and the link is up.

The server listens on TCP port 569. Nothing is written to the socket before the
first record in either direction: no greeting, no `COM` exchange, no version
string.

Bring-up then runs per section 3.1.3, with the client's type 4 frame first on
the wire:

```
Client -> record, control type 4      immediately on connect, unprompted
Server -> record, control type 3      transport parameters
Client -> record, control type 1      connection request
Server -> record, control type 1      echo
```

The client sends its type 4 without waiting and sends nothing further until the
parameters arrive.

#### 4.3.4 Sending a Pipe Unit

Emit `TotalLength = 3 + len(message)`, the command byte of section 4.3.1, and the
message. Records
are written back to back with nothing between them, and one record is written
whole before the next begins (section 2.1.1).

#### 4.3.5 Receiving Records

1. Append received bytes to ReceiveBuffer.
2. While ReceiveBuffer holds at least 3 bytes, read TotalLength. If it is below
   3, the stream is unsynchronized: it carries no marker to resynchronize on, so
   the receiver MUST treat the connection as failed. If ReceiveBuffer holds
   fewer than TotalLength bytes, stop and wait for more.
3. Take the record: byte 2 is the command-byte echo and is discarded, bytes 3
   through TotalLength-1 are the pipe message. Deliver the message as one pipe
   unit and remove the record from ReceiveBuffer.

A record may arrive split across any number of TCP segments, and several records
may arrive in one. Neither is visible above this section.

## 5 Protocol Examples

Every byte sequence in this section is complete and consistent with the rules
above, and every length and check field in it can be recomputed from sections 2
and 4. Bytes are shown in hexadecimal, most significant nibble first.

Examples that show a bare pipe message or a bare host block are
transport-independent: they are what the pipe layer submits, and each binding
frames them per section 4.

| Example | Shows |
|---|---|
| 5.1 | Bring-up, on both bindings (sections 3.1.3, 4.2.5.1, 4.3.3) |
| 5.2 | Pipe open, response, and interface table (sections 2.2.3.2, 2.2.3.3, 2.2.6) |
| 5.3 | A call and a static reply, with both variable-field size forms (sections 2.2.7, 2.2.8) |
| 5.4 | Dynamic replies, the two-block sequence terminator, and a cancel (sections 2.2.8.4, 2.2.9) |
| 5.5 | Fragmentation and reassembly (sections 4.2.8, 4.2.9) |
| 5.6 | A chunked field and its deferred reply (sections 2.2.7.2, 3.3.5.3) |
| 5.7 | An error reply (section 2.2.8.5) |
| 5.8 | A pipe close, and the record byte that echoes its command (sections 2.2.4, 4.3.1) |

### 5.1 Bring-Up

The four connection-level messages of section 3.1.3.

Server, transport parameters:

```
ff ff 03 00 04 00 00 00 04 00 00 10 00 00 00 01
00 00 00 58 02 00 00
```

| Bytes | Meaning |
|---|---|
| `ff ff` | Routing: control frame |
| `03` | Control type 3, transport parameters |
| `00 04 00 00` | PacketSize 1024 |
| `00 04 00 00` | MaxBytes 1024 |
| `10 00 00 00` | WindowSize 16 |
| `01 00 00 00` | AckBehind 1 |
| `58 02 00 00` | AckTimeout 600 ms |

The message is 23 bytes, so no KeepAlive follows.

Client, connection established:

```
ff ff 04
```

Client, connection request, and the server's echo of it:

```
ff ff 01 01 00 00 00
```

#### 5.1.1 Over Select

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
| `e0` | Pipe header: reassembly index 0, Continuation, LastData |
| `17 00` | ContentLength 23 |
| `ff ff 03 …` | The pipe message |
| `1b 32` | The WindowSize byte 0x10, escape-encoded |
| `38 c9 9a 7e` | Check field |
| `0d` | Terminator |

Client to server, connection established. This follows the framing handshake and
precedes the transport parameters, so the client has received no data packet yet
and acknowledges 0:

```
80 80 e0 03 00 ff ff 04 fc 03 18 a0 0d
```

Server to client, the echo of the connection request, under its own sequence and
acknowledgment numbers:

```
81 82 e0 07 00 ff ff 01 01 00 00 00 40 60 42 74 0d
```

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

Acknowledgment packets flow between the data packets shown here and are omitted
from the remaining examples. A peer that has a data packet of its own to send
carries the acknowledgment in it instead (section 4.2.7 step 9).

#### 5.1.2 Over Straight

The client connects and sends, with nothing before it:

```
06 00 00 ff ff 04
```

| Bytes | Meaning |
|---|---|
| `06 00` | TotalLength 6 |
| `00` | Cmd: a control frame has no command byte to echo |
| `ff ff 04` | The pipe message |

Server to client, transport parameters:

```
1a 00 00 ff ff 03 00 04 00 00 00 04 00 00 10 00
00 00 01 00 00 00 58 02 00 00
```

The WindowSize byte 0x10 is transmitted as itself: this binding does not escape.

Client to server, then server to client, the connection request and its echo:

```
0a 00 00 ff ff 01 01 00 00 00
```

### 5.2 Opening a Service Pipe

Client to server, the pipe message opening service `LOGSRV` version 6 on pipe 3:

```
00 00 00 00 03 00 4c 4f 47 53 52 56 00 55 00 06
00 00 00
```

| Bytes | Meaning |
|---|---|
| `00 00` | Routing: pipe-open request |
| `00 00` | Reserved |
| `03 00` | PipeIndex 3 |
| `4c 4f 47 53 52 56 00` | ServiceName "LOGSRV" |
| `55 00` | Parameter "U" |
| `06 00 00 00` | Version 6 |

Server to client, the pipe-open response:

```
03 00 01 00 03 00 00 00
```

| Bytes | Meaning |
|---|---|
| `03 00` | Routing: pipe 3 |
| `01 00` | Command: pipe opened |
| `03 00` | ServerPipeIndex 3 |
| `00 00` | Status: success |

Server to client, the interface table for the service. The bytes below are the
first three of its ten records, so they are a prefix of the pipe message and not
a complete one; the whole message is 175 bytes, its 5-byte head and 10 records
of 17 bytes each. Which GUIDs those records carry, and which identifier each
receives, are the service's and the server's to choose (section 2.2.6):

```
03 00
00 00 00
b6 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46 01
b7 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46 02
b8 8b 02 00 00 00 00 00 c0 00 00 00 00 00 00 46 03
```

| Bytes | Meaning |
|---|---|
| `03 00` | Routing: pipe 3 |
| `00 00 00` | Class 0, Method 0, RequestId 0 |
| `b6 8b 02 00 …46` | GUID 00028BB6-0000-0000-C000-000000000046 |
| `01` | Interface identifier 1 for that GUID |

Over Select, the request and the response:

```
82 82 e0 13 00 00 00 00 00 03 00 4c 4f 47 53 52     client, seq 2, ack 2
56 00 55 00 06 00 00 00 8e b6 6e 20 0d

82 83 e3 08 00 03 00 01 00 03 00 00 00 b6 d3 09     server, seq 2, ack 3
2d 0d
```

Over Straight, the same two:

```
16 00 00 00 00 00 00 03 00 4c 4f 47 53 52 56 00     client
55 00 06 00 00 00

0b 00 00 03 00 01 00 03 00 00 00                    server
```

### 5.3 A Call with a Static Reply

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

Server to client, the reply: seven dwords, end-of-static, and a 16-byte variable
field, as a pipe message on pipe 3:

```
03 00
06 00 00
83 00 00 00 00 83 00 00 00 00 83 00 00 00 00 83
00 00 00 00 83 00 00 00 00 83 00 00 00 00 83 00
00 00 00
87
84 90 00 00 00 00 00 00 00 00 00 00 00 00 00 00
00 00
```

| Bytes | Meaning |
|---|---|
| `03 00` | Routing: pipe 3 |
| `06 00 00` | Class 6, Method 0, RequestId 0, repeated from the call |
| `83 00 00 00 00` × 7 | Seven dword fields, answering the seven `83` descriptors |
| `87` | End of static section |
| `84 90 …` | Variable field, size byte 0x90: bit 7 set, length 16 |

Over Select, in one packet of 70 wire bytes:

```
84 84 e3 3b 00 03 00 06 00 00 83 00 00 00 00 83
00 00 00 00 83 00 00 00 00 83 00 00 00 00 83 00
00 00 00 83 00 00 00 00 83 00 00 00 00 87 84 1b
35 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
00 2b 11 96 ab 0d
```

| Bytes | Meaning |
|---|---|
| `84 84` | Seq 4, Ack 4 |
| `e3` | Pipe header: reassembly index 3, Continuation, LastData |
| `3b 00` | ContentLength 59, no byte of which needs escaping |
| `1b 35` | The size byte 0x90, escape-encoded |
| `2b 11 96 ab` | Check field |

Over Straight, in one record of 62 bytes:

```
3e 00 00 03 00 06 00 00 83 00 00 00 00 83 00 00
00 00 83 00 00 00 00 83 00 00 00 00 83 00 00 00
00 83 00 00 00 00 83 00 00 00 00 87 84 90 00 00
00 00 00 00 00 00 00 00 00 00 00 00 00 00
```

### 5.4 A Call with a Dynamic Reply

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

**Cancelling the sequence instead.** A caller that abandons the sequence on
interface 6, method 2, request identifier 9 sends a call block whose body is the
single byte 0x0F, and the peer answers `87 88` on the same three values:

```
06 02 09 0f            client to server, cancel
06 02 09 87 88         server to client, the answer
```

The answer is sent whether or not the server still has that sequence (section
2.2.9). After it, the server sends nothing further on request identifier 9 and
the identifier is free for reuse.

### 5.5 A Reply of 2060 Bytes on Pipe 5

The pipe message is the same on both bindings: the routing value `05 00`
followed by a 2058-byte host block, 2060 bytes in all.

Over Select with PacketSize 1024, starting at sequence 10, three packets carry
it. No byte of this message is one of the reserved values of section 4.2.1.1, so
the encoding expands nothing and each fragment is as large as PacketSize allows:

| Packet | Wire length | Seq | Pipe header | LastData | Encoded payload | Message bytes |
|---|---|---|---|---|---|---|
| 1 | 1024 | 10 | `a5` | no | 1017 | 1014 |
| 2 | 1024 | 11 | `a5` | no | 1017 | 1016 |
| 3 | 38 | 12 | `e5` | yes | 31 | 30 |

The payload ceiling is `PacketSize − 2 − 4 − 1` = 1017: the Seq and Ack bytes,
the check field, and the terminator. Packet 1 spends 3 of its 1017 on the frame
header and the two-byte ContentLength, packets 2 and 3 spend 1 on the frame
header alone.

First bytes of packet 1:

```
8a 84 a5 0c 08 05 00 03 …
```

| Bytes | Meaning |
|---|---|
| `8a` | Seq 10 |
| `84` | Ack 4 |
| `a5` | Pipe header: reassembly index 5, Continuation, LastData clear |
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
| `a5` | Pipe header: reassembly index 5, Continuation, LastData clear |
| `ce …` | Message bytes. No length field: only frame 1 carries one |

The receiver's reassembly state (section 4.2.3) across the three packets:

| After | Context | Buffer | BytesOutstanding | Delivered |
|---|---|---|---|---|
| start | — | empty | 0 | — |
| packet 1 | 5 | 1014 bytes | 1046 | — |
| packet 2 | 5 | 2030 bytes | 30 | — |
| packet 3 | 5 | 2060 bytes | 0 | the 2060-byte pipe unit; its routing value names pipe 5 |

The message completes when BytesOutstanding reaches zero, not when LastData
arrives. Had the sender cleared LastData on packet 3 by mistake, the message
would still be delivered on the same byte; had it set LastData on packet 1, a
receiver that ended the message there would have joined all three fragments and
the next message into one buffer.

All three frames must carry the same reassembly index, since it is what binds
them to one context; which of the sixteen it is does not matter. A client
sending the same message would write `a0`, `a0`, `e0` — index 0 — throughout
(section 3.1.5.1). The receiver takes the routing value `05 00` inside the
completed message as the destination.

Over Straight it is one record of 2063 bytes:

```
0f 08 00 05 00 03 …
```

| Bytes | Meaning |
|---|---|
| `0f 08` | TotalLength 2063 |
| `05` | Pipe 5 |
| `05 00` | Routing: pipe 5 |
| `03 …` | Host block |

### 5.6 A Call Carrying a Chunked Field

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

The server holds the call open while the frames arrive. When the 0xE7 frame
completes the 1500 bytes it runs the method and answers, with a reply carrying
one field per descriptor the call declared — a dword and a variable field, for
the `83 84` above:

```
0b 04 07 83 00 00 00 00 87 84 80
```

| Bytes | Meaning |
|---|---|
| `0b 04 07` | Class 11, Method 4, RequestId 7, repeated from the call |
| `83 00 00 00 00` | Status dword, answering the `83` descriptor |
| `87` | End of static section |
| `84 80` | Variable field, size byte 0x80: bit 7 set, length 0 |

Had a frame been lost and the stream stopped at 1496 bytes, the same call would
have been answered with the error field instead (section 3.3.5.3):

```
0b 04 07 8f 09 00 00 e0 87
```

### 5.7 An Error Reply

The call of section 5.3 — interface 6, method 0, request identifier 0, on
pipe 3 — answered with 0xE0000004, the method not being registered on that
interface. The reply body is the error field and the end-of-static tag, and
none of the eight fields the call's descriptors declared:

```
03 00
06 00 00
8f 04 00 00 e0
87
```

| Bytes | Meaning |
|---|---|
| `03 00` | Routing: pipe 3 |
| `06 00 00` | Class 6, Method 0, RequestId 0, repeated from the call |
| `8f` | Error field |
| `04 00 00 e0` | 0xE0000004, little-endian |
| `87` | End of static section |

The pipe message is 11 bytes. Over Select, at sequence 5:

```
85 85 e3 1b 33 00 03 00 06 00 00 8f 04 00 00 e0
87 ee b4 47 3b 0d
```

| Bytes | Meaning |
|---|---|
| `85 85` | Seq 5, Ack 5 |
| `e3` | Pipe header: reassembly index 3, Continuation, LastData |
| `1b 33 00` | ContentLength 11. The low byte 0x0B is a reserved value and is escape-encoded |
| `ee b4 47 3b` | Check field |

Over Straight, one record of 14 bytes:

```
0e 00 00 03 00 06 00 00 8f 04 00 00 e0 87
```

### 5.8 Closing a Pipe

The client closes pipe 3. The pipe message is three bytes: the routing value of
the pipe being closed, and the command byte.

```
03 00 01
```

| Bytes | Meaning |
|---|---|
| `03 00` | Routing: pipe 3 |
| `01` | Pipe close |

Over Select, at sequence 4, on reassembly index 0:

```
84 85 e0 03 00 03 00 01 e8 a1 fa 29 0d
```

| Bytes | Meaning |
|---|---|
| `84 85` | Seq 4, Ack 5 |
| `e0` | Pipe header: reassembly index 0, Continuation, LastData |
| `03 00` | ContentLength 3 |
| `03 00 01` | The pipe message: close pipe 3 |
| `e8 a1 fa 29` | Check field |

Over Straight, one record of 6 bytes:

```
06 00 01 03 00 01
```

| Bytes | Meaning |
|---|---|
| `06 00` | TotalLength 6 |
| `01` | Cmd: the close command byte, echoed from the message (section 4.3.1) |
| `03 00 01` | The pipe message: close pipe 3 |

The `01` at byte 2 is the command byte, not pipe 1. It is the one place the
Straight field is ever non-zero, and it is why a receiver that read the field as
a pipe number would close pipe 1 here, leave pipe 3 open, and strand every call
outstanding on it.

No response is sent to a close on either binding. If this was the last service
pipe on the connection, the server then releases the link (section 3.3.5.4).

## 6 Security

### 6.1 Security Considerations for Implementers

MOS RPC provides no confidentiality, no integrity protection against
modification, and no authentication of either peer, on any binding. The Select
check field detects corruption, not tampering: it is a CRC with a published
polynomial, and any party that can modify the byte stream can recompute it. TCP
offers no more.

Credentials carried as call parameters are visible to anyone with access to the
link. Deployments that need confidentiality must obtain it below this protocol.

Authentication is a service concern. Identity established by one service applies
to the whole connection (section 3.3.1), so a server MUST treat every pipe on a
connection as carrying the same principal, and MUST NOT infer identity from the
pipe a call arrives on.

### 6.2 Index of Security Parameters

An implementation MUST bound the following, all of which are attacker-controlled
in a message and each of which otherwise allows a peer to force unbounded
allocation or a non-terminating loop. Every peer enforces every bound in this
section, on both bindings.

Protocol. Where a bound is defined elsewhere this table cites it and adds only
what a receiver does when the bound is breached:

| Parameter | Bound | Breach |
|---|---|---|
| Pipe message length | 65,532 bytes (section 2.2.2) | Discard the message. |
| Variable field size | 32,767 bytes, the largest the size encoding expresses (section 2.2.1.3), and MUST NOT exceed the bytes remaining in the host block | Error 0xE0000007. |
| Dynamic message | 16,384 bytes per message (section 2.2.8.4) | Discard the message and fail the call. |
| Send parameters | 16 per request (section 2.2.7.1) | Error 0xE0000007. |
| Receive descriptors | 16 per request (section 2.2.7.3) | Error 0xE0000007. |
| Chunked field length | 16,777,216 bytes declared, per stream | Error 0xE0000009; the stream is never opened. |
| Decompressed field size | An implementation-chosen ceiling on the bytes a 0x44 or 0x45 field expands to (section 2.2.7.4) | Error 0xE0000007 for 0x44, 0xE0000009 for 0x45; the partial output is discarded. |
| Concurrent streams | One per stream identifier, 256 per connection | Error 0xE0000009 on the call that would exceed it. |
| Outstanding calls | 256 per pipe | Discard the call. |
| Pipes | 15 per connection (section 2.2.2) | Discard the pipe-open request (section 3.3.5.1). |
| Control frame type-specific field | 4,096 bytes | Discard the frame; do not echo it (section 2.2.3.1.1). |
| ServiceName, Parameter | 256 bytes each, including the NUL | Discard the pipe-open request. |

The chunked-field bound is a protocol ceiling, not a service one. A service MAY
refuse a smaller field, and MUST answer 0xE0000009 when it does rather than
discard the stream silently; a client whose upload is refused otherwise waits
out a timeout the protocol does not define. Two implementations that pick
different service ceilings disagree about which uploads work, so a service that
means to be portable states its ceiling in its own contract.

The decompressed-size bound has no protocol value because no length on the wire
describes it: a 0x44 or 0x45 field declares only its compressed size (section
2.2.7.4), and the expansion ratio is bounded by nothing. A receiver MUST
therefore count output bytes as it decompresses and stop at its own ceiling,
rather than allocate against the declared length or run the codec to
completion first.

Select binding:

| Parameter | Bound | Breach |
|---|---|---|
| Packet length | PacketSize | Discard the packet. |
| ContentLength | 65,532 bytes (section 2.2.2) | Discard the message. |
| Message reassembly | A frame that would carry more bytes than BytesOutstanding is truncated to it | Never allowed to overrun. |
| Reassembly buffer lifetime | Sixteen partial messages per connection, one per reassembly index (section 4.2.3), each released when its message completes | A peer that opens a message on every index and finishes none holds sixteen buffers for the life of the connection. |

Straight binding:

| Parameter | Bound | Breach |
|---|---|---|
| TotalLength | A record is buffered only up to TotalLength | A TotalLength below 3 fails the connection (section 4.3.5). |

## 7 Appendix A: Constants

This appendix summarizes values defined in sections 2, 4 and 6. It is not
normative: where it and a defining section differ, the defining section governs.

### 7.1 Protocol

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
| 0xE0 - 0xFF | Reserved; only the two below are assigned |
| 0xE6 | Stream frame, more follow |
| 0xE7 | Stream frame, last |

**Request tags**

| Tag | Meaning |
|---|---|
| 0x01 | byte |
| 0x02 | word |
| 0x03 | dword |
| 0x04 | variable |
| 0x44 | variable, compressed |
| 0x05 | chunked field reference |
| 0x45 | chunked field reference, compressed |
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

See section 2.2.8.5.

**Fixed limits**

Defined in section 6.2, which also gives what a receiver does on a breach.

| Name | Value |
|---|---|
| Pipe close command | 0x01 |
| Iterator cancel body | 0x0F |

### 7.2 Select Binding

**Packet framing**

| Name | Value |
|---|---|
| Terminator | 0x0D |
| Escape byte | 0x1B |
| Acknowledgment packet type | 0x41 |
| Negative acknowledgment packet type | 0x42 |
| Reserved values requiring the transform of section 4.2.1.2 | 0x8D, 0x90, 0x8B |
| Transform mask | 0xC0 |
| Check polynomial | 0x248EF9BE |
| Check mask | OR 0x60 on 0x1B, 0x0D, 0x10, 0x0B, 0x8D, 0x90, 0x8B |
| Minimum packet length | 7 bytes |
| Link initialization reply | `COM` 0x0D |
| Link initialization timeout | 20 s |
| Retransmission limit | 12 |
| Sequence space | 128 |

**Telnet negotiation on a bridged link** (section 4.2.5.2)

| Name | Value |
|---|---|
| IAC | 0xFF |
| WILL | 0xFB |
| WONT | 0xFC |
| DO | 0xFD |
| DONT | 0xFE |
| SB | 0xFA |
| SE | 0xF0 |

**Pipe header bits**

| Name | Mask |
|---|---|
| ReassemblyIndex | 0x0F |
| HasLength | 0x10 |
| Continuation | 0x20 |
| LastData | 0x40 |
| Always set | 0x80 |

**Negotiated parameters**

Defined in section 2.2.3.1.2, which gives their defaults and how each peer
applies them.

### 7.3 Straight Binding

| Name | Value |
|---|---|
| Record header length | 3 bytes |
| Maximum record | 65,535 bytes |
| TCP port | 569 |

## 8 Appendix B: Implementation Interface

This appendix is non-normative. It restates the types, constants, state, and
operation boundaries implied by sections 2 through 4 and section 6 in a
compact declaration-only form. It is intended as an implementation scaffold,
not as an application SDK: blocking calls, iterators, callbacks, allocation
policy, threading, and service-specific method contracts are deliberately not
defined here.

The declarations are split into four layers. `wire` describes normalized wire
syntax and codecs; `transport` describes the byte-stream dependency and the two
bindings; `protocol.state` mirrors the abstract data models in section 3; and
`protocol.ops` names the state transitions and higher-layer triggers described
there. A target language may merge layers, but doing so must not change their
semantics.

Notation: `u8`, `u16`, and `u32` are unsigned integers; `bytes<N>` is an exact
byte array; `bytes<=N` is a bounded byte vector; `bytes` is implementation-owned
storage with no wire-size implication; `list<T,N>` and `map<K,V,N>` have at most
`N` entries; `array<T,N>` has exactly `N`; `optional<T>`, `result<T,E>`, and
`union` have their usual meanings; `handle X`
is opaque implementation state; and `inout` means the operation may mutate the
object. Multi-byte wire integers are little-endian unless explicitly marked
otherwise.

```mosidl
module mos_rpc;

/* ======================================================================== */
/* wire: scalar types and protocol constants                                */
/* ======================================================================== */

namespace wire {

type Byte                 = u8;
type Word                 = u16;
type Dword                = u32;
type Vli                  = u32 where value <= 0x3fffffff;
type FieldSize            = u16 where value <= 32767; /* senders emit <= 32766 */
type RoutingValue         = u16;
type ReassemblyIndex      = u8  where value <= 15;
type ServicePipeIndex     = u8  where 1 <= value && value <= 15;
type WireServicePipeIndex = u16 where 1 <= value && value <= 15;
type InterfaceId          = u8  where 0x01 <= value && value <= 0xdf;
type MethodId             = u8;
type RequestId            = Vli where value <= 0x3ffffffe;
type StreamId             = u8;
type SequenceNumber       = u8  where value <= 127;
type ServiceVersion       = u32;
type DurationMs           = u32;
type ChunkedLength        = u32 where value <= 16777216;

type ServiceName          = ascii_z<=256; /* bound includes terminating NUL */
type ServiceParameter     = ascii_z<=256; /* bound includes terminating NUL */
type ConnectionLog        = bytes<=4096 where 4 <= len(value);
type PipeMessageBytes     = bytes<=65532;
type PipeContentBytes     = bytes<=65530;
type HostBlockBytes       = bytes<=65530;
type CallBodyBytes        = bytes<=65527; /* encoder also enforces total host block <= 65530 */
type InlineVariableBytes  = bytes<=32766;
type LogicalVariableBytes = bytes<=16777216;
type DynamicBytes         = bytes<=16384;
type StreamFrameBytes     = bytes<=65528;

const MAX_PIPES: u32                       = 16;
const MAX_SERVICE_PIPES: u32               = 15;
const MAX_PIPE_MESSAGE: u32                = 65532;
const MAX_VARIABLE_FIELD: u32              = 32767;
const MAX_INLINE_VARIABLE_FIELD: u32       = 32766; /* largest a sender encodes inline */
const MAX_DYNAMIC_MESSAGE: u32             = 16384;
const MAX_SEND_PARAMETERS: u32             = 16;
const MAX_RECEIVE_DESCRIPTORS: u32         = 16;
const MAX_CHUNKED_FIELD: u32                = 16777216;
const MAX_CONCURRENT_STREAMS: u32          = 256;
const MAX_OUTSTANDING_CALLS_PER_PIPE: u32  = 256;
const MAX_CONTROL_TYPE_DATA: u32           = 4096;
const MAX_SERVICE_STRING: u32              = 256;
const MAX_VLI: u32                         = 0x3fffffff;
const MAX_REQUEST_ID: u32                  = 0x3ffffffe; /* largest a sender encodes as a VLI */

const ROUTING_PIPE_OPEN: u16               = 0x0000;
const ROUTING_PIPE_DATA_MIN: u16           = 0x0001;
const ROUTING_PIPE_DATA_MAX: u16           = 0x000f;
const ROUTING_CONTROL: u16                 = 0xffff;

const PIPE_OPEN_RESERVED: u16              = 0x0000;
const PIPE_OPEN_COMMAND: u16               = 0x0001;
const PIPE_OPEN_STATUS: u16                = 0x0000;
const PIPE_CLOSE_COMMAND: u8               = 0x01;
const ITERATOR_CANCEL_BODY: u8             = 0x0f;

const CONTROL_CONNECTION_REQUEST: u8       = 0x01;
const CONTROL_TRANSPORT_PARAMETERS: u8     = 0x03;
const CONTROL_CONNECTION_ESTABLISHED: u8   = 0x04;

const CLASS_INTERFACE_TABLE: u8            = 0x00;
const CLASS_INTERFACE_MIN: u8              = 0x01;
const CLASS_INTERFACE_MAX: u8              = 0xdf;
const CLASS_RESERVED_MIN: u8               = 0xe0;
const CLASS_RESERVED_MAX: u8               = 0xff;
const CLASS_STREAM_MORE: u8                = 0xe6;
const CLASS_STREAM_FINAL: u8               = 0xe7;

const SEND_BYTE: u8                        = 0x01;
const SEND_WORD: u8                        = 0x02;
const SEND_DWORD: u8                       = 0x03;
const SEND_VARIABLE: u8                    = 0x04;
const SEND_VARIABLE_COMPRESSED: u8         = 0x44;
const SEND_CHUNKED: u8                     = 0x05;
const SEND_CHUNKED_COMPRESSED: u8          = 0x45;
const SEND_COMPRESSED_BIT: u8              = 0x40;
const COMPRESSED_BLOCK_MAX_PLAIN: u32      = 32768;

const RECEIVE_BYTE: u8                     = 0x81;
const RECEIVE_WORD: u8                     = 0x82;
const RECEIVE_DWORD: u8                    = 0x83;
const RECEIVE_VARIABLE: u8                 = 0x84;
const RECEIVE_DYNAMIC: u8                  = 0x85;

const REPLY_DYNAMIC_MORE: u8               = 0x85;
const REPLY_CALL_COMPLETE: u8              = 0x86;
const REPLY_END_STATIC: u8                 = 0x87;
const REPLY_MESSAGE_COMPLETE: u8           = 0x88;
const REPLY_ERROR: u8                      = 0x8f;

const DEFAULT_PACKET_SIZE: u32             = 1024;
const DEFAULT_MAX_BYTES: u32               = 1024;
const DEFAULT_WINDOW_SIZE: u32             = 16;
const DEFAULT_ACK_BEHIND: u32              = 1;
const DEFAULT_ACK_TIMEOUT_MS: u32          = 600;

struct Guid {
    data1: u32;
    data2: u16;
    data3: u16;
    data4: bytes<8>;
}

struct TransportParameters {
    packet_size: u32;
    max_bytes: u32;
    window_size: u32;
    ack_behind: u32;
    ack_timeout_ms: u32;
    keep_alive_ms: optional<u32>;
}

/* Complete shared error namespace exposed by implementations. */
enum ErrorCode : u32 {
    PARAMETER_TYPE_MISMATCH = 0xe0000001;
    INVALID_HOST_BLOCK      = 0xe0000002;
    PARAMETER_ADD_OR_SEND   = 0xe0000003;
    METHOD_NOT_REGISTERED   = 0xe0000004;
    ALLOCATION_FAILED       = 0xe0000005;
    INTERNAL_ERROR          = 0xe0000006;
    INVALID_PARAMETER       = 0xe0000007;
    SEND_ERROR              = 0xe0000008;
    CHUNK_RECEIVE_ERROR     = 0xe0000009;
    SERVICE_REFUSED         = 0xe000000a;
}

/* The only values legal in a server-emitted 0x8f field. */
enum WireErrorCode : u32 {
    METHOD_NOT_REGISTERED = 0xe0000004;
    INTERNAL_ERROR        = 0xe0000006;
    INVALID_PARAMETER     = 0xe0000007;
    CHUNK_RECEIVE_ERROR   = 0xe0000009;
    SERVICE_REFUSED       = 0xe000000a;
}

/* Internal syntax failures have no wire numeric value of their own. */
enum CodecError {
    TRUNCATED;
    MALFORMED;
    NONCANONICAL_FIELD_SIZE;
    UNASSIGNED_SEND_TAG;
    BOUND_EXCEEDED;
    TYPE_MISMATCH;
}

/* A complete pipe unit at the abstract transport boundary. The binding's own
   framing field is not part of it; see SelectPipeFrame and StraightRecord. */
struct PipeUnit {
    message: PipeMessageBytes;
}

/* The raw pipe-layer unit. Interpretation of content depends on routing/state. */
struct PipeMessage {
    routing: RoutingValue;
    content: PipeContentBytes;
}

/* Normalized control-frame forms. */
union ControlFrame {
    connection_request(ConnectionLog);
    transport_parameters(TransportParameters);
    connection_established(void);
}

/* Reserved/command/status fields are normalized away by these declarations. */
struct PipeOpenRequest {
    pipe_index: ServicePipeIndex;
    service_name: ServiceName;
    parameter: ServiceParameter;
    version: ServiceVersion;
}

struct PipeOpenResponse {
    pipe_index: ServicePipeIndex;
}

struct InterfaceRecord {
    guid: Guid;
    interface_id: InterfaceId;
}

struct InterfaceTable {
    records: list<InterfaceRecord,223>
        where unique(records.guid) && unique(records.interface_id);
}

enum ChunkedTag : u8 {
    PLAIN      = 0x05;
    COMPRESSED = 0x45;
}

/* length always counts wire bytes: compressed ones when tag is COMPRESSED. */
struct ChunkedReference {
    tag: ChunkedTag;
    stream_id: StreamId;
    length: ChunkedLength;
}

enum ReceiveType : u8 {
    BYTE     = 0x81;
    WORD     = 0x82;
    DWORD    = 0x83;
    VARIABLE = 0x84;
    DYNAMIC  = 0x85;
}

/* Wire representation: inline vs chunked and plain vs compressed are visible here.
   Compressed payloads are carried verbatim; the block codec is out of scope. */
union WireSendValue {
    byte(Byte);
    word(Word);
    dword(Dword);
    variable(InlineVariableBytes);
    variable_compressed(InlineVariableBytes);
    chunked_reference(ChunkedReference);
}

/* Application/logical representation: chunking is not a parameter type. */
union SendArgument {
    byte(Byte);
    word(Word);
    dword(Dword);
    variable(LogicalVariableBytes);
}

union RequestElement {
    send(WireSendValue);
    receive(ReceiveType);
}

/* Senders place sends first; receivers must also accept interleaving. */
struct RequestParameters {
    elements: list<RequestElement,32>
        where count(elements.send) <= 16 && count(elements.receive) <= 16;
}

union RequestBody {
    parameters(RequestParameters);
    iterator_cancel(void);
}

union StaticValue {
    byte(Byte);
    word(Word);
    dword(Dword);
    variable(InlineVariableBytes);
}

union DynamicPart {
    none;
    more(DynamicBytes);             /* tag 0x85 */
    call_complete(DynamicBytes);    /* tag 0x86 */
    message_complete(DynamicBytes); /* tag 0x88 */
}

struct ReplySuccessBody {
    static_values: list<StaticValue,16>;
    post_static_values: list<InlineVariableBytes,16>;
    dynamic: DynamicPart;
}

union ReplyBody {
    success(ReplySuccessBody);
    error(WireErrorCode);
}

/* Class 0x01..0xdf has one opaque body; request/reply meaning is contextual. */
struct CallBlock {
    interface_id: InterfaceId;
    method_id: MethodId;
    request_id: RequestId;
    body: CallBodyBytes;
}

enum StreamFrameKind : u8 {
    MORE  = 0xe6;
    FINAL = 0xe7;
}

struct StreamFrame {
    kind: StreamFrameKind;
    stream_id: StreamId;
    bytes: StreamFrameBytes;
}

union HostBlock {
    interface_table(InterfaceTable); /* class 0x00, method 0, request-id 0 */
    call(CallBlock);                 /* class 0x01..0xdf */
    stream(StreamFrame);             /* class 0xe6 or 0xe7 */
}

union DecodeOrDiscard<T> {
    value(T);
    discard;
}

/* Request parsing has protocol-visible failure semantics. */
enum RequestWireErrorCode : u32 {
    INVALID_PARAMETER   = 0xe0000007;
    CHUNK_RECEIVE_ERROR = 0xe0000009;
}

union RequestParseResult {
    value(RequestBody);
    wire_error(RequestWireErrorCode);
}

struct Decoded<T> {
    value: T;
    consumed: u32;
}

interface CodecOps {
    fn encode_vli(value: Vli) -> bytes<=4;
    fn decode_vli(input: bytes) -> result<Decoded<Vli>,CodecError>;

    fn encode_field_size(value: FieldSize) -> bytes<=2;
    fn decode_field_size(input: bytes) -> result<Decoded<FieldSize>,CodecError>;

    fn encode_guid(value: Guid) -> bytes<16>;
    fn decode_guid(input: bytes<16>) -> Guid;

    fn encode_pipe_message(value: PipeMessage) -> PipeMessageBytes;
    fn decode_pipe_message(input: PipeMessageBytes)
        -> DecodeOrDiscard<PipeMessage>;

    fn encode_control_frame(value: ControlFrame) -> PipeMessage;
    fn decode_control_frame(value: PipeMessage)
        -> DecodeOrDiscard<ControlFrame>;

    fn encode_pipe_open_request(value: PipeOpenRequest) -> PipeMessage;
    fn decode_pipe_open_request(value: PipeMessage)
        -> DecodeOrDiscard<PipeOpenRequest>;

    fn encode_pipe_open_response(value: PipeOpenResponse) -> PipeMessage;
    fn decode_pipe_open_response(value: PipeMessage)
        -> DecodeOrDiscard<PipeOpenResponse>;

    fn encode_pipe_close(pipe: ServicePipeIndex) -> PipeMessage;

    fn encode_host_block(value: HostBlock) -> result<HostBlockBytes,ErrorCode>;
    fn decode_host_block(input: HostBlockBytes)
        -> result<HostBlock,CodecError>;

    fn encode_request_body(value: RequestBody) -> result<CallBodyBytes,ErrorCode>;
    fn decode_request_body(input: CallBodyBytes)
        -> RequestParseResult;

    fn encode_reply_body(value: ReplyBody) -> result<CallBodyBytes,ErrorCode>;
    fn decode_reply_body(
        input: CallBodyBytes,
        expected: list<ReceiveType,16>,
        first_block: bool
    ) -> result<ReplyBody,ErrorCode>;
}

} /* namespace wire */

/* ======================================================================== */
/* transport: link dependency and binding implementations                    */
/* ======================================================================== */

namespace transport {

using wire.*;

/* Link errors are supplied by the host environment, not MOS RPC. */
handle LinkFailure;
handle ByteStream;

union LinkRead {
    data(bytes);
    closed;
}

interface ByteStreamOps {
    fn read(stream: ByteStream, max_bytes: u32) -> result<LinkRead,LinkFailure>;
    fn write_all(stream: ByteStream, data: bytes) -> result<void,LinkFailure>;
    fn close(stream: ByteStream) -> void;
}

enum PeerRole {
    CLIENT;
    SERVER;
}

enum BindingKind {
    SELECT;
    STRAIGHT;
}

/* No numeric MOS RPC error is assigned to a binding/link failure. */
handle TransportFailure;
handle TransportBinding;

enum SelectLinkMode {
    RAW;
    TELNET_BRIDGED;
}

struct BindingConfig {
    local_maxima: TransportParameters;
    select_link_mode: optional<SelectLinkMode>;
    select_server_init_timeout_ms: optional<u32>;
    select_server_init_noise_limit: optional<u32>;
}

interface TransportBindingOps {
    fn initialize(
        kind: BindingKind,
        role: PeerRole,
        stream: ByteStream,
        config: BindingConfig
    ) -> result<TransportBinding,TransportFailure>;

    fn local_parameters(binding: TransportBinding) -> TransportParameters;

    fn apply_peer_parameters(
        binding: TransportBinding,
        parameters: TransportParameters
    ) -> void;

    fn send(binding: TransportBinding, unit: PipeUnit)
        -> result<void,TransportFailure>;

    fn receive(binding: TransportBinding)
        -> result<PipeUnit,TransportFailure>;

    /* Select: clear the one connection-wide reassembly context. Straight: no-op. */
    fn discard_partial(binding: TransportBinding) -> void;

    fn close(binding: TransportBinding) -> void;
}

/* ---- Select binding ----------------------------------------------------- */

const SELECT_TERMINATOR: u8              = 0x0d;
const SELECT_ESCAPE: u8                  = 0x1b;
const SELECT_ACK_PACKET: u8              = 0x41;
const SELECT_NAK_PACKET: u8              = 0x42;
const SELECT_TRANSFORM_MASK: u8          = 0xc0;
const SELECT_CHECK_POLYNOMIAL: u32       = 0x248ef9be;
const SELECT_CRC32_REFLECTED_POLY: u32   = 0xedb88320;
const SELECT_CHECK_MASK_OR: u8           = 0x60;
const SELECT_MIN_PACKET_LENGTH: u32      = 7;
const SELECT_INIT_TIMEOUT_MS: u32        = 20000;
const SELECT_RETRANSMISSION_LIMIT: u32   = 12;
const SELECT_SEQUENCE_SPACE: u32         = 128;
const SELECT_PIPE_INDEX_MASK: u8         = 0x0f;
const SELECT_HAS_LENGTH: u8              = 0x10;
const SELECT_CONTINUATION: u8            = 0x20;
const SELECT_LAST_DATA: u8               = 0x40;
const SELECT_HEADER_ALWAYS_SET: u8       = 0x80;
const SELECT_INIT_REPLY: bytes<4>         = "COM\r";

const SELECT_TELNET_IAC: u8              = 0xff;
const SELECT_TELNET_WILL: u8             = 0xfb;
const SELECT_TELNET_WONT: u8             = 0xfc;
const SELECT_TELNET_DO: u8               = 0xfd;
const SELECT_TELNET_DONT: u8             = 0xfe;
const SELECT_TELNET_SB: u8               = 0xfa;
const SELECT_TELNET_SE: u8               = 0xf0;

const SELECT_ESCAPE_RAW: bytes<7>         = [0x1b,0x0d,0x10,0x0b,0x8d,0x90,0x8b];
const SELECT_ESCAPE_CODE: bytes<7>        = [0x30,0x31,0x32,0x33,0x34,0x35,0x36];
const SELECT_TRANSFORM_RAW: bytes<3>      = [0x8d,0x90,0x8b];
const SELECT_TRANSFORM_WIRE: bytes<3>     = [0x4d,0x50,0x4b];
const SELECT_CHECK_MASK_BYTES: bytes<7>   = [0x1b,0x0d,0x10,0x0b,0x8d,0x90,0x8b];

type SelectPacketWireBytes = bytes; /* one complete packet including terminator */
type SelectPayloadBytes    = bytes; /* decoded payload */

struct SelectDataPacket {
    sequence: SequenceNumber;
    acknowledgment: SequenceNumber;
    payload: SelectPayloadBytes; /* decoded payload */
}

struct SelectAckPacket {
    acknowledgment: SequenceNumber;
}

struct SelectNakPacket {
    last_intact_sequence: SequenceNumber;
}

union SelectPacket {
    data(SelectDataPacket);
    acknowledgment(SelectAckPacket);
    negative_acknowledgment(SelectNakPacket);
}

struct SelectHasLengthFrame {
    reassembly_index: ReassemblyIndex;
    last_data: bool;
    content: bytes<=127;
}

struct SelectContinuationFrame {
    reassembly_index: ReassemblyIndex;
    last_data: bool;
    content: bytes;
}

union SelectPipeFrame {
    has_length(SelectHasLengthFrame);
    continuation(SelectContinuationFrame);
}

struct SelectOutstandingPacket {
    sequence: SequenceNumber;
    wire_bytes: SelectPacketWireBytes;
}

struct SelectActiveReassembly {
    buffer: PipeMessageBytes;
    bytes_outstanding: u32 where 0 < value && value <= 65532;
}

union SelectReassemblyState {
    idle;
    active(SelectActiveReassembly);
}

struct SelectState {
    send_sequence: SequenceNumber;
    expected_sequence: SequenceNumber;
    unacknowledged_packets: list<SelectOutstandingPacket>;
    effective_window: u32;
    received_since_ack: u32;
    reassembly: array<SelectReassemblyState,16>; /* indexed by reassembly index */
    parameters: TransportParameters;
    retransmission_timer_running: bool;
    acknowledgment_timer_running: bool;
    keep_alive_timer_running: bool;
    consecutive_retransmission_expiries: u32;
}


enum SelectTimerKind {
    RETRANSMISSION;
    ACKNOWLEDGMENT;
    KEEP_ALIVE;
}

struct SelectTimerArm {
    timer: SelectTimerKind;
    interval_ms: DurationMs;
}

union SelectReceiveAction {
    write_packet(SelectPacketWireBytes);
    deliver_pipe_unit(PipeUnit);
    start_timer(SelectTimerArm);
    stop_timer(SelectTimerKind);
    reset_timer(SelectTimerArm);
    fail_connection(void);
}

struct SelectTelnetFilterState {
    pending: bytes;
}

struct SelectTelnetFilterResult {
    data: bytes;
    replies: list<bytes>;
}

interface SelectOps {
    fn initialize_link(
        role: PeerRole,
        stream: ByteStream,
        config: BindingConfig
    ) -> result<void,TransportFailure>;

    fn transform(value: u8) -> u8;
    fn reverse_transform(value: u8) -> u8;
    fn escape_encode(input: bytes) -> bytes;
    fn escape_decode(input: bytes) -> result<bytes,CodecError>;
    fn check(input: bytes) -> u32;
    fn mask_check(value: u32) -> bytes<4>;

    fn filter_telnet_initial_bytes(
        state: inout SelectTelnetFilterState,
        input: bytes
    ) -> SelectTelnetFilterResult;

    fn encode_packet(packet: SelectPacket, packet_size: u32)
        -> result<SelectPacketWireBytes,TransportFailure>;
    fn decode_packet(input: SelectPacketWireBytes, packet_size: u32)
        -> DecodeOrDiscard<SelectPacket>;

    /* Conforming senders always emit Continuation form; decoders accept both forms. */
    fn encode_frame(frame: SelectContinuationFrame) -> bytes;
    fn decode_next_frame(
        input: bytes,
        bytes_outstanding: u32
    ) -> result<Decoded<SelectPipeFrame>,CodecError>;

    fn fragment_pipe_unit(
        state: inout SelectState,
        unit: PipeUnit
    ) -> result<list<SelectPacketWireBytes>,TransportFailure>;

    fn receive_packet(
        state: inout SelectState,
        input: SelectPacketWireBytes
    ) -> list<SelectReceiveAction>;

    fn retransmission_timeout(
        state: inout SelectState
    ) -> list<SelectReceiveAction>;

    fn acknowledgment_timeout(
        state: inout SelectState
    ) -> list<SelectReceiveAction>;

    fn keep_alive_timeout(
        state: inout SelectState
    ) -> list<SelectReceiveAction>;

    fn discard_partial(state: inout SelectState) -> void;
}

/* ---- Straight binding --------------------------------------------------- */

const STRAIGHT_TCP_PORT: u16              = 569;
const STRAIGHT_RECORD_HEADER_LENGTH: u32  = 3;
const STRAIGHT_MAX_RECORD: u32            = 65535;

type StraightTotalLength = u16 where 3 <= value;

/* total_length is derived as 3 + len(message), not caller-supplied state. */
struct StraightRecord {
    command_echo: Byte; /* the message's command byte, or 0; never routing */
    message: PipeMessageBytes;
}

struct StraightState {
    receive_buffer: bytes<=65535;
}

union StraightDecodeResult {
    record(Decoded<StraightRecord>);
    need_more;
    connection_failure; /* TotalLength < 3 */
}

interface StraightOps {
    fn initialize_link(
        role: PeerRole,
        stream: ByteStream,
        config: BindingConfig
    ) -> result<void,TransportFailure>; /* no wire exchange; config is inert */

    fn encode_record(record: StraightRecord) -> bytes<=65535;
    fn decode_record(input: bytes) -> StraightDecodeResult;

    fn feed(
        state: inout StraightState,
        input: bytes
    ) -> result<list<PipeUnit>,TransportFailure>;
}

} /* namespace transport */

/* ======================================================================== */
/* protocol.state: state explicitly named by sections 3.1, 3.2, and 3.3     */
/* ======================================================================== */

namespace protocol.state {

using wire.*;
using transport.TransportBinding;

enum PipePhase {
    CLOSED;
    OPENING;
    OPEN;
}

enum ClientBringUpPhase {
    WAIT_TRANSPORT_PARAMETERS;
    WAIT_CONNECTION_ECHO;
    READY;
}

enum CallerKind {
    SINGLE_RESULT;
    SEQUENCE;
}

struct CallAddress {
    interface_id: InterfaceId;
    method_id: MethodId;
    request_id: RequestId;
}

struct ClientOutstandingCall {
    address: CallAddress;
    receive_types: list<ReceiveType,16>;
    caller_kind: CallerKind;
    first_reply_block: bool;
    dynamic_buffer: DynamicBytes;
    cancel_pending: bool;
}

struct ClientOpeningPipe {
    service_name: ServiceName;
    parameter: ServiceParameter;
    version: ServiceVersion;
    required_interfaces: list<Guid,223>;
}

struct ClientOpenPipe {
    service_name: ServiceName;
    version: ServiceVersion;
    interface_table_received: bool;
    interfaces: map<Guid,InterfaceId,223>;
    outstanding_calls: map<RequestId,ClientOutstandingCall,256>;
    next_request_id: RequestId;
}

union ClientPipeSlot {
    closed;
    opening(ClientOpeningPipe);
    open(ClientOpenPipe);
}

struct ClientState {
    binding: TransportBinding;
    transport_parameters: optional<TransportParameters>;
    bring_up: ClientBringUpPhase;
    connection_log: ConnectionLog;
    pipes: array<ClientPipeSlot,16>; /* by pipe number; entry 0 unused */
    next_stream_id: StreamId;
}

handle SessionState;
handle ServiceBinding;

struct ServerOutstandingCall {
    address: CallAddress;
    receive_types: list<ReceiveType,16>;
    referenced_streams: list<StreamId,16>;
}

struct OpenStream {
    stream_id: StreamId;
    pipe_index: ServicePipeIndex;
    call_address: CallAddress;
    declared_length: ChunkedLength;
    received_length: u32; /* the spec defines short-final handling, not an overrun rule */
    final_seen: bool;
}

struct ServerOpenPipe {
    service: optional<ServiceBinding>;
    interfaces: map<InterfaceId,Guid,223>;
    outstanding_calls: map<RequestId,ServerOutstandingCall,256>;
}

union ServerPipeSlot {
    closed;
    opening;
    open(ServerOpenPipe);
}

struct ServerState {
    binding: TransportBinding;
    transport_parameters: TransportParameters;
    pipes: array<ServerPipeSlot,16>; /* by pipe number; entry 0 unused */
    session: optional<SessionState>;
    open_streams: map<StreamId,OpenStream,256>;
    had_service_pipe: bool;
}

} /* namespace protocol.state */

/* ======================================================================== */
/* protocol.ops: protocol transitions, not convenience/runtime APIs          */
/* ======================================================================== */

namespace protocol.ops {

using wire.*;
using transport.*;
using protocol.state.*;

/* Routing must use the two-byte routing value, never a binding framing field. */

union PipeDispatch {
    control(ControlFrame);
    pipe_open_request(PipeOpenRequest);
    pipe_open_response(PipeOpenResponse);
    pipe_close(ServicePipeIndex);
    host_block(struct { pipe_index: ServicePipeIndex; block: HostBlock; });
    discard;
}

/* Parameters seen by a server method; chunked inputs are references, not buffers.
   `compressed` is true when the tag carried bit 6 and the bytes are still deflated. */
union ReceivedArgument {
    byte(Byte);
    word(Word);
    dword(Dword);
    variable(struct { bytes: InlineVariableBytes; compressed: bool; });
    chunked_reference(ChunkedReference);
}

struct ServerInvocation {
    pipe_index: ServicePipeIndex;
    service: ServiceBinding;
    interface_guid: Guid;
    session: optional<SessionState>;
    address: CallAddress;
    arguments: list<ReceivedArgument,16>;
    receive_types: list<ReceiveType,16>;
}

/* Effects are protocol-visible consequences; their execution policy is external. */
union ProtocolEffect {
    send_pipe_unit(PipeUnit);
    apply_transport_parameters(TransportParameters);
    discard_transport_partial(void);
    close_transport(void);

    interface_table_received(struct {
        pipe_index: ServicePipeIndex;
        table: InterfaceTable;
    });

    reply_fields(struct {
        pipe_index: ServicePipeIndex;
        address: CallAddress;
        static_values: list<StaticValue,16>;
        post_static_values: list<InlineVariableBytes,16>;
    });

    single_complete(struct {
        pipe_index: ServicePipeIndex;
        address: CallAddress;
        dynamic_bytes: optional<DynamicBytes>;
    });

    sequence_message(struct {
        pipe_index: ServicePipeIndex;
        address: CallAddress;
        bytes: DynamicBytes;
    });

    sequence_end(struct {
        pipe_index: ServicePipeIndex;
        address: CallAddress;
    });

    call_failed(struct {
        pipe_index: ServicePipeIndex;
        address: CallAddress;
        error: ErrorCode;
    });

    invoke_server_method(ServerInvocation);

    stream_bytes(struct {
        stream_id: StreamId;
        bytes: bytes;
    });

    stream_complete(struct {
        stream_id: StreamId;
    });

    iterator_cancelled(struct {
        pipe_index: ServicePipeIndex;
        address: CallAddress;
    });
}

/* The protocol does not define service contracts; a registry is implementation-owned. */
handle ServiceRegistry;

/* Logical call description used before inline/chunked wire encoding is chosen. */
struct ClientCallSpec {
    interface_guid: Guid;
    method_id: MethodId;
    arguments: list<SendArgument,16>;
    receive_types: list<ReceiveType,16>;
    caller_kind: CallerKind;
}

struct TransitionResult<T> {
    value: T;
    effects: list<ProtocolEffect>;
}

interface ProtocolOps {
    /* Take a free reassembly index for a new outbound message, per section
       3.1.5.1. Fails when all sixteen are in flight; the sender waits. */
    fn allocate_reassembly_index(
        in_flight: array<bool,16>
    ) -> optional<ReassemblyIndex>;

    fn dispatch_pipe_message(
        role: PeerRole,
        pipe_phases: array<PipePhase,16>, /* by pipe number; entry 0 unused */
        message: PipeMessage
    ) -> PipeDispatch;

    fn client_pipe_phases(state: ClientState) -> array<PipePhase,16>;
    fn server_pipe_phases(state: ServerState) -> array<PipePhase,16>;

    /* Both begin operations receive an already initialized binding. */
    fn client_begin(
        binding: TransportBinding,
        connection_log: ConnectionLog
    ) -> result<TransitionResult<ClientState>,ErrorCode>;

    fn server_begin(
        binding: TransportBinding
    ) -> result<TransitionResult<ServerState>,ErrorCode>;

    fn client_receive_unit(
        state: inout ClientState,
        unit: PipeUnit
    ) -> list<ProtocolEffect>;

    fn server_receive_unit(
        state: inout ServerState,
        unit: PipeUnit,
        services: ServiceRegistry
    ) -> list<ProtocolEffect>;

    fn client_open_service(
        state: inout ClientState,
        service_name: ServiceName,
        parameter: ServiceParameter,
        version: ServiceVersion,
        required_interfaces: list<Guid,223>
    ) -> result<TransitionResult<ServicePipeIndex>,ErrorCode>;

    fn client_issue_call(
        state: inout ClientState,
        pipe_index: ServicePipeIndex,
        spec: ClientCallSpec
    ) -> result<TransitionResult<CallAddress>,ErrorCode>;

    fn client_cancel_sequence(
        state: inout ClientState,
        pipe_index: ServicePipeIndex,
        address: CallAddress
    ) -> result<TransitionResult<void>,ErrorCode>;

    fn client_close_pipe(
        state: inout ClientState,
        pipe_index: ServicePipeIndex
    ) -> list<ProtocolEffect>;

    fn server_close_pipe(
        state: inout ServerState,
        pipe_index: ServicePipeIndex
    ) -> list<ProtocolEffect>;

    fn client_link_failed(state: inout ClientState) -> list<ProtocolEffect>;
    fn server_link_failed(state: inout ServerState) -> list<ProtocolEffect>;

    fn server_set_session_state(
        state: inout ServerState,
        session: SessionState
    ) -> void;

    /* ReplyBody.error can carry only WireErrorCode, never an arbitrary ErrorCode. */
    fn server_send_reply_block(
        state: inout ServerState,
        pipe_index: ServicePipeIndex,
        address: CallAddress,
        body: ReplyBody
    ) -> result<list<ProtocolEffect>,ErrorCode>;

    fn server_send_no_reply(
        state: inout ServerState,
        pipe_index: ServicePipeIndex,
        address: CallAddress
    ) -> list<ProtocolEffect>;

    fn server_push_sequence_message(
        state: inout ServerState,
        pipe_index: ServicePipeIndex,
        address: CallAddress,
        bytes: DynamicBytes
    ) -> result<list<ProtocolEffect>,ErrorCode>;

    fn server_end_sequence(
        state: inout ServerState,
        pipe_index: ServicePipeIndex,
        address: CallAddress
    ) -> result<list<ProtocolEffect>,ErrorCode>;
}

} /* namespace protocol.ops */
```

The declarations intentionally leave several choices to an implementation
because the protocol does: storage for chunked fields, the threshold at which a
short variable argument is nevertheless sent chunked, call timeouts, thread or
event-loop structure, service registration, and method contracts. In
particular, a logical variable argument is represented once as `SendArgument`
and may be encoded inline or by `ChunkedReference`; `WireSendValue` exists only
at the wire-codec boundary.
