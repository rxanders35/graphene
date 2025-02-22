use raft::RawNode;

mod build;

#[tokio::main]
async fn main() {
    #[allow(unused)]
    let cfg = raft::Config::new(0);
    #[allow(unused)]
    let logger = unimplemented!();
    #[allow(unused)]
    let raft_log = unimplemented!();
    // let node = RawNode::new(&cfg, raft_log, logger);
}
