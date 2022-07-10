# simple-lsm-db

simple database based on LSM, previously on [simple-hash-db](https://github.com/0x822a5b87/simple-hash-db) I created a simple hash-index database which contains many problem such as performance and efficiency.However, I'm not familiar with cpp, so I decided to create a new project using `Golang` to solve this problem.

## reference

- [从零开始写数据库：500行代码实现 LSM 数据库](https://zhuanlan.zhihu.com/p/374535126)
- [TinyKvStore](https://github.com/x-hansong/TinyKvStore)
- [SSTable](https://www.scylladb.com/glossary/sstable/)

## required

![go1.18](https://badgen.net/badge/go/1.18/red?icon=github)

## feature

- `SSTable` use SSTable instead of single file for higher compact performance
- `sparse index` use sparse index instead of memory-hash index for
  - lower memory
  - reduce IO cost
  - range query
- `WAL` use write ahead log to improve the reliability
- `cache` use cache to improve query performance

## sequence graph

### set

```mermaid
sequenceDiagram
    participant client;
    participant server;
    participant index;
    participant immutableIndex;
    participant wal;
    participant tmpWal;
    participant ssTables;
    participant indexLock;

    client ->> server: set()
    server ->>+ indexLock: writeLock().lock()
    server ->> wal: write command length
    server ->> wal: write commaond
    server ->> index: put
    alt should persist?
        server ->> server: shoudld persist?
        note over index,immutableIndex: switch
        wal ->> wal: close()
        note over wal,tmpWal: switch
        immutableIndex ->> ssTables: storeToSsTables()
        immutableIndex ->> immutableIndex: delete()
    end
    indexLock ->>- server: writeLock().unlock()
    server ->> client: set result.
```

### get

```mermaid
sequenceDiagram
    participant client;
    participant server;
    participant index;
    participant immutableIndex;
    participant ssTables;
    participant indexLock;

    client ->> server: get()
    server ->>+ indexLock: readLock().lock()
    server ->> index: get()
    alt not null
        index->>server: value
    else is null
        server ->> immutableIndex: get()
        alt not null
            immutableIndex ->> server: value
        else is null
            server ->> ssTables: get()
            alt is null
                ssTables ->> server: null
            else not null
                ssTables ->> server: value
            end
        end
    end
    indexLock ->>- server: readLock().lock()
    alt is null
        server->>client: null
    else not null
        alt is set
            server->>client: value
        else is rm
            server->>client: null
        end
    end
```

### .table

![.table structure](./resources/_table.jpg)

### SsTable#initFromIndex

```mermaid
sequenceDiagram
    participant tableMetaInfo
    participant index
    participant tableFile
    participant sparseIndex

    tableMetaInfo ->> tableMetaInfo: setDataStart

    loop index.values()
        index ->> tableFile : write part data
        index ->> sparseIndex : write sparse index
    end

    tableMetaInfo ->> tableFile: getFilePointer
    tableFile ->> tableMetaInfo: file pointer
    tableMetaInfo ->> tableMetaInfo: setDataLen
    tableMetaInfo ->> tableMetaInfo: setIndexStart
    index ->> tableFile: write index
    tableMetaInfo ->> tableFile: getFilePointer
    tableFile ->> tableMetaInfo: file pointer
    tableMetaInfo ->> tableMetaInfo : setIndexLen
    tableMetaInfo ->> tableFile: save table metadata info
```

### SsTable#restoreFromFile

```mermaid
sequenceDiagram
    participant tableMetaInfo
    participant tableFile
    participant sparseIndex

    tableMetaInfo ->> tableFile: read from file
    sparseIndex ->> tableFile: read from file
```

### SsTable#query

```mermaid

```

## code analysis for TinyKvStore

### code structure

1. we store `Command` instead of store raw data directly, so we do not need `GetCommand`. When we invoke `get`, we search the last `SetCommand` in order of `memory index`, `persisting SSTable`， `persisted SSTable`.

```
src/main/java/com/xiaohansong/kvstore/
        ├── model
        │     ├── Position.java
        │     ├── command
        │     │     ├── AbstractCommand.java                     abstract command with CommandTypeEnum
        │     │     ├── Command.java                             command interface
        │     │     ├── CommandTypeEnum.java                     command type
        │     │     ├── RmCommand.java                           rm command
        │     │     └── SetCommand.java                          set command
        │     └── sstable
        │         ├── SsTable.java                               
        │         └── TableMetaInfo.java
        ├── service
        │     ├── KvStore.java
        │     └── LsmKvStore.java
        └── utils
              ├── ConvertUtil.java                               convert JSONObject to Command
              └── LoggerUtil.java                                logger util
```

## code analysis for simple-lsm-db
