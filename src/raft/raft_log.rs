use openraft::{
    entry::{FromAppData, RaftEntry, RaftPayload},
    BasicNode, LogId, NodeId, RaftLogId,
};

#[derive(Debug)]
pub enum RaftCommand {
    UploadRequest,
    UploadResponse,
}

#[derive(Debug)]
pub struct LogEntry {
    command: RaftCommand,
    idx: u64,
    term: u64,
}

impl RaftLogId<u64> for LogEntry {
    fn get_log_id(&self) -> &LogId<u64> {
        todo!();
    }

    fn set_log_id(&mut self, log_id: &LogId<u64>) {
        todo!();
    }
}

impl RaftPayload<u64, BasicNode> for LogEntry {
    fn get_membership(&self) -> Option<&openraft::Membership<u64, BasicNode>> {
        todo!();
    }

    fn is_blank(&self) -> bool {
        todo!();
    }
}

impl RaftEntry<u64, BasicNode> for LogEntry {
    fn new_blank(log_id: openraft::LogId<u64>) -> Self {
        todo!();
    }

    fn new_membership(
        log_id: openraft::LogId<u64>,
        m: openraft::Membership<u64, BasicNode>,
    ) -> Self {
        todo!();
    }
}

impl FromAppData<RaftCommand> for LogEntry {
    fn from_app_data(t: RaftCommand) -> Self {
        todo!()
    }
}

impl From<RaftCommand> for LogEntry {
    fn from(command: RaftCommand) -> Self {
        Self {
            command,
            idx: 0,
            term: 0,
        }
    }
}

impl std::fmt::Display for LogEntry {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "[cmd: {} - idx: {} - term: {}]",
            match self.command {
                RaftCommand::UploadRequest => "UploadRequest",
                RaftCommand::UploadResponse => "UploadResponse",
            },
            self.idx,
            self.term
        )
    }
}
