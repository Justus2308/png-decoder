# PNG de-/encoder from/to BMP
## About
This is a simple terminal-based application to decode files from PNG to BMP and back.

It supports standard PNG filtering and will soon support Adam7-Interlacing. All crtical PNG chunks are fully supported, but ancilliary chunks are currently being ignored except for a warning about their existence in the terminal output.

PNG images can be truecolour with/without alpha or paletted. Greyscale is not supported. The only supported bit depth is 8 bit.
BMP images with BITMAPINFOHEADER, V4INFOHEADER and V5INFOHEADER are supported. An alpha channel is only supported for the latter two. Only uncompressed BMP images with either 1, 3 or 4 bpp (bytes per pixel) can be decoded.

All BMP images created with this application will have either a BITMAPINFOHEADER if they don't have an alpha channel or a V5INFOHEADER if they do.

I chose BMP as a format to convert from/to because it saves images as lossless, uncompressed bitmaps in its base form.

## Usage
To compile the code, simply run build.sh which will create a 'run' executable. The application runs in a terminal window and will ask you to enter one of four commands:
- encode "path" -alpha=true/false inter=true/false ; encode BMP to PNG
- decode "path" -alpha=true/false ; decode PNG to BMP
- help ; print help for all commands
- quit ; quit the application

All flags are optional, their standard values are: alpha=true, inter=false. The flag values will be reset to their standard value after each encoding/decoding process.

Interlacing is currently not supported so setting the flag to true will not do anything.

## Implementation
The entire application is written in Google's Golang. Go's specialty is concurrency using goroutines, which will be implemented into this application in the future. I think the PNG filtering process expecially could really benefit from concurrency. Unfortunately, reading and writing PNG files is for the most part an inherently sequential process, as it often relies on already de-/encoded data in both the filtering and the compression process.

### Encoder
#### 1. Decoding the BMP
The encoder starts by decoding the given BMP's magic numbers and header. My code for this is basically a modified version of Go's native image/bmp reader and will determine the dimensions and colour depth of the image. If the BMP header is a V4INFOHEADER or a V5INFOHEADER and the image has 4 bpp, it is assumed that the image supports transparency. 3 bpp is treated as a truecolour image witout transparency and 1 bpp as a paletted image. When decoding the BMP header it is important to note that is uses the little-endian format, in contrast to PNG which uses big-endian.

BMP pixels are represented as BGR or BGRA/BGRX, which must first be converted to RGB or RGBA for PNG.

Also, most BMP files are top-down, which is signalized by a negative image height, so the scanline order has to be reversed in that case.

#### 2. Filtering the bitmap
For better compression results, image data steams are first filtered line by line through one of five methods, which is chosed by calculating the filtered line's minimum absoulte difference. The filter method with the lowest score will be used.

The minimum absoulte difference is the most commonly used heuristic for determining the best filter method and obviously way less expensive than compressing every line with every filter type and choosing the one which uses the least disk space. It is calculated by sorting the line's bytes by value and summing up the lowest differences of each byte to its neighbour. My implementation is based on the pnglib encoder, as there is no official documentation on specific heuristics to determine the best filter type. Information on this topic is pretty scarce in general, probably because there is no official standard.

There are five filter types defined:
- 0 none: filt(x) = orig(x)
- 1 sub: filt(x) = orig(x) - orig(x-bpp)
- 2 up: filt(x) = orig(x) - prior(x)
- 3 average: filt(x) = orig(x) - floor((orig(x-bpp) + prior(x)) / 2)
- 4 paeth: filt(x) = orig(x) - PaethPredictor(orig(x-bpp), prior(x), prior(x-bpp))

filt: filtered scanline ; orig: original scanline ; prior: prior scanline

The PaethPredictor function is a PNG specific implementation of the Paeth algorithm, as seen [here](https://www.w3.org/TR/2003/REC-PNG-20031110/#9Filter-type-4-Paeth "official documentation").

Filters always work byte-by-byte. They use bytes of the current and of the prior scanline for their calculations, so the previous, unfiltered scanline must always be saved until the next one is fully filtered. The sequential reading of scanlines is implemented with Go's reader interface, which goes through the bmp file scanline by scanline.

After choosing the supposedly best filter type for each line, a filter type byte is prepended to the filtered scanline. This is implemented with as little new memory allocations as possible by appending one byte to the scanline, copying it onto itself with one byte offset and assigning the filter type byte to scanline[0].

#### 3. Creating data chunks
A PNG is made up of at least three critical data chunks: the IHDR (image header), one or multiple IDAT (image data) chunks and an IEND (image end) chunk. If the image is paletted, an additional PLTE (palette) chunk after the IHDR and before any IDAT chunks is mandatory. There are many other chunk types, but those are considered ancilliary chunks and are not supported by this application.

Multiple IDAT chunks must be consecutive and their image data will be concatenated on decoding.
A chunk is made up of four main data fields: the length, its magic numbers, the data field and a CRC32 checksum generated to the IEEE standard.

They contain what one would expect for the most part, except for the IDAT chunk. Its data field is not the raw image data stream, but it is compressed first by the zlib deflate algorithm. This is implemented line-by-line, which is not very efficient for small images, but it scales well. To keep this application from having to store the entire deflated data stream until the compression is completed, every scanline is encoded as a seperate IDAT chunk and instantly written to the target file.

#### 4. Creating the PNG file
Every PNG file has to start with the PNG-specific magic numbers. Afterwards, the first chunk has to be an IHDR chunk. If the image is paletted, there has to be one PLTE chunk present before the first IDAT chunk. The number of IDAT chunks is unlimited, as long as they are consecutive. The last chunk has to be an IEND chunk, any data after the IEND chunk will not be read by decoders.

### Decoder
#### 1. Decoding the PNG
First, the decoder checks if the magic numbers are correct. Then it checks if the first chunk is an IHDR by looking at its (fixed) length and its magic numbers. Finally it checks, if its CRC32 checksum is correct. Then it reads its data, like image dimensions and colour depth.

If the image is paletted, the next critical chunk has to be a PLTE chunk. There can only be one PLTE chunk per image.
For every chunk the decoder will check the chunk type first. If the chunk is ancilliary, it will issue a warning to the user and ignore the chunk. If the chunk is critical, but not either an IDAT or IEND chunk, it will return an error. Whether a chunk is critical is determined by looking at its magic number's first byte. If it is in uppercase, it's critical, if not, it is ancilliary. This catches e.g. multiple IHDR or PLTE chunks, or even unknown, not officialy specified critical chunks.

For ancilliary chunks, the decoder also checks the magic number's second byte. If it is uppercase, it is a public chunk, which means it is defined by the official PNG standard. If not, it is a private, unofficial chunk. The ancilliary chunk type and whether it's public or private will be part of the warning issued to the user.

When the IEND chunk occurs, the decoder will stop reading the PNG file.
All concatenated IDAT chunks have to be kept in memory simultaneously, as their deflated data stream has its own checksum which is generated from the entire data stream and needs to be checked by the zlib inflate algorithm.

#### 2. Reconstructing the image data stream
After all IDAT chunks are decompressed, the filters are reversed line-by-line by simply applying the proper reconstruction functions based on the scanline's filter byte.

For every filter type, there is also a reconstruction function:
- 0 none: recon(x) = filt(x)
- 1 sub: recon(x) = filt(x) + recon(x-bpp)
- 2 up: recon(x) = filt(x) + prior(x)
- 3 average: recon(x) = filt(x) + floor((recon(x-bpp) + prior(x)) / 2)
- 4 paeth: recon(x) = filt(x) + PaethPredictor(recon(x-bpp), prior(x), prior(x-bpp))

recon: reconstructed scanline ; filt: filtered scanline ; prior: prior, reconstructed scanline

The filter type bytes are ignored by reconstruction and will be trimmed from the final result.

#### 3. Creating the BMP file
Every BMP file has to start with the BMP-specific magic numbers. Afterwards, the appropiate header is added (BMPINFOHEADER for images without transparency, V5INFOHEADER for images with a working alpha channel). After the header, the raw image data stream is simply appended.

## Known Issues / WIP
Decoding paletted images is currently untested.


Interlacing will be implemented in the future.
