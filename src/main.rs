use raft::RawNode;

mod build;

#[tokio::main]
async fn main() {
    let cfg = raft::Config::new(0);
    let logger = unimplemented!();
    #[allow(unused)]
    let raft_log = unimplemented!();
    let node = RawNode::new(&cfg, raft_log, logger);
}
