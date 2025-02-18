//impl Send/OptionalSend + 'static
pub struct Snapshot {
    pub curr_files: Vec<String>,
    pub time_stamp: u64,
}
