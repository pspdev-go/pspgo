# MVP design notes

The verified legacy flow in `pspsdk-go/build-sample.sh` was:

1. force `GOTOOLCHAIN=local`, compile `example` with `GOMIPS=softfloat`,
   `-gc=psp`, and TinyGo's built-in `psp` target; scheduler selection is
   inherited from that target rather than overridden by the driver;
2. inspect surviving undefined symbols after TinyGo dead-code elimination;
3. scan every archive under `$PSPDEV/psp/sdk/lib`, select defining members,
   and recursively include their dependencies;
4. let `psp-cmake` compile `main.c`, package markers and ABI helpers with PSP
   GCC, link all archives in a group, then invoke `mksfoex` and `pack-pbp`.

The old Python resolver mixed archive discovery policy, bridge knowledge and
CMake serialization. The MVP ports graph resolution to Go and moves bridge
metadata into a typed registry. Generated CMake is now merely a backend.

Stage failures are labeled `compile`, `dependency`, `configure`, or
`link/package`. This makes failures actionable without requiring users to
understand the whole chain.

The next architectural seam is a `Frontend` interface producing object files
and ABI metadata. TinyGo is its first implementation; an LLVM pass, ABI
rewriter or independent backend can later implement the same contract without
changing dependency resolution and packaging.
