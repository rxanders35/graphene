use super::raft_log_entry::{LogEntry, RaftCommand};
use openraft::{declare_raft_types, BasicNode};
use std::io::Cursor;

declare_raft_types!(
    pub TypeConfig:
        D = RaftCommand,
        R = RaftCommand,
        NodeId = u64,
        Node = BasicNode, //impl my own
        Entry = LogEntry,
        SnapshotData = Cursor<Vec<u8>>, //impl my own
        AsyncRuntime = openraft::TokioRuntime,
);
