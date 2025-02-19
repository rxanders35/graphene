use openraft::storage::{LogFlushed, RaftLogStorage};
use openraft::{LogId, OptionalSend, RaftLogReader, RaftTypeConfig, StorageError};

use sled::Db;

use super::raft_log_entry::LogEntry;

pub struct LogStorage {
    db: Db,
    entries: Vec<LogEntry>,
}

pub struct LogStorageReader {
    store: LogStorage,
}

impl<C: RaftTypeConfig> RaftLogReader<C> for LogStorageReader {
    async fn limited_get_log_entries(
        &mut self,
        start: u64,
        end: u64,
    ) -> Result<Vec<C::Entry>, StorageError<C::NodeId>> {
        todo!();
    }

    async fn try_get_log_entries<
        RB: std::ops::RangeBounds<u64> + Clone + std::fmt::Debug + OptionalSend,
    >(
        &mut self,
        range: RB,
    ) -> Result<Vec<C::Entry>, StorageError<C::NodeId>> {
        todo!();
    }
}

impl<C: RaftTypeConfig> RaftLogStorage<C> for LogStorage {
    type LogReader = LogStorageReader;

    async fn get_log_state(&mut self) -> Result<openraft::LogState<C>, StorageError<C::NodeId>> {
        todo!()
    }

    async fn get_log_reader(&mut self) -> Self::LogReader {
        todo!()
    }

    async fn save_vote(
        &mut self,
        vote: &openraft::Vote<C::NodeId>,
    ) -> Result<(), StorageError<C::NodeId>> {
        todo!()
    }

    async fn save_committed(
        &mut self,
        _committed: Option<LogId<<C as RaftTypeConfig>::NodeId>>,
    ) -> Result<(), StorageError<<C as RaftTypeConfig>::NodeId>> {
        todo!()
    }

    async fn read_committed(
        &mut self,
    ) -> Result<
        Option<LogId<<C as RaftTypeConfig>::NodeId>>,
        StorageError<<C as RaftTypeConfig>::NodeId>,
    > {
        todo!()
    }

    async fn read_vote(
        &mut self,
    ) -> Result<
        Option<openraft::Vote<<C as RaftTypeConfig>::NodeId>>,
        StorageError<<C as RaftTypeConfig>::NodeId>,
    > {
        todo!()
    }

    async fn append<I>(
        &mut self,
        entries: I,
        callback: LogFlushed<C>,
    ) -> Result<(), StorageError<<C as RaftTypeConfig>::NodeId>>
    where
        I: IntoIterator<Item = <C as RaftTypeConfig>::Entry> + OptionalSend,
        I::IntoIter: OptionalSend,
    {
        todo!()
    }

    async fn truncate(
        &mut self,
        log_id: LogId<<C as RaftTypeConfig>::NodeId>,
    ) -> Result<(), StorageError<<C as RaftTypeConfig>::NodeId>> {
        todo!()
    }

    async fn purge(
        &mut self,
        log_id: LogId<<C as RaftTypeConfig>::NodeId>,
    ) -> Result<(), StorageError<<C as RaftTypeConfig>::NodeId>> {
        todo!();
    }
}
