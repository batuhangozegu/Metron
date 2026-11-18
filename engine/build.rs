// Generates the gRPC server code from the shared contract in ../proto at
// build time; the generated code is never committed.
fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Use the bundled protoc so no system install is needed (locally or in CI).
    // SAFETY: build scripts are single-threaded; nothing else reads the env yet.
    unsafe {
        std::env::set_var("PROTOC", protoc_bin_vendored::protoc_bin_path()?);
    }

    // The client is generated too, for the over-the-wire tests.
    tonic_prost_build::configure().compile_protos(
        &[
            "../proto/metron/engine/v1/engine.proto",
            "../proto/metron/engine/v1/risk.proto",
        ],
        &["../proto"],
    )?;
    Ok(())
}
