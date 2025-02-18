use crate::raft::raft_fabric::Snapshot;
use crate::raft::raft_log::{LogEntry, RaftCommand};
use openraft::{declare_raft_types, BasicNode, RaftTypeConfig};
use std::io::Cursor;

declare_raft_types!(
    pub TypeConfig:
        D = RaftCommand,
        R = RaftCommand,
        NodeId = u64,
        Node = openraft::BasicNode, //impl my own
        Entry = LogEntry,
        SnapshotData = Cursor<Vec<u8>>, //impl my own
        AsyncRuntime = openraft::TokioRuntime,
);
