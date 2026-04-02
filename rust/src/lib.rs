use percent_encoding::percent_decode_str;
use std::ffi::{CStr, CString};
use std::io::Read;
use std::os::raw::c_char;
use std::ptr;
use std::sync::atomic::{AtomicBool, Ordering};
use wl_clipboard_rs::copy::{MimeType as CopyMimeType, Options as CopyOptions, Source};
use wl_clipboard_rs::paste::{
    get_contents, ClipboardType, Error as ClipboardError, MimeType, Seat,
};

static FOREGROUND: AtomicBool = AtomicBool::new(false);

fn get_foreground() -> bool {
    FOREGROUND.load(Ordering::SeqCst)
}

#[repr(C)]
pub struct WlClipString {
    pub ptr: *mut c_char,
    pub len: usize,
    pub error: *mut c_char,
}

#[repr(C)]
pub struct WlClipBytes {
    pub ptr: *mut u8,
    pub len: usize,
    pub error: *mut c_char,
}

#[repr(C)]
pub struct WlClipInt {
    pub value: i32,
    pub error: *mut c_char,
}

fn ok_string(s: String) -> WlClipString {
    match CString::new(s) {
        Ok(c) => {
            let len = c.as_bytes().len();
            WlClipString {
                ptr: c.into_raw(),
                len,
                error: ptr::null_mut(),
            }
        }
        Err(_) => err_string("string contains null byte"),
    }
}

fn err_string(msg: &str) -> WlClipString {
    WlClipString {
        ptr: ptr::null_mut(),
        len: 0,
        error: CString::new(msg).unwrap().into_raw(),
    }
}

#[allow(dead_code)]
fn ok_bytes(data: Vec<u8>) -> WlClipBytes {
    let mut b = data.into_boxed_slice();
    let len = b.len();
    let ptr = b.as_mut_ptr();
    std::mem::forget(b);
    WlClipBytes {
        ptr,
        len,
        error: ptr::null_mut(),
    }
}

#[allow(dead_code)]
fn err_bytes(msg: &str) -> WlClipBytes {
    WlClipBytes {
        ptr: ptr::null_mut(),
        len: 0,
        error: CString::new(msg).unwrap().into_raw(),
    }
}

fn ok_int(value: i32) -> WlClipInt {
    WlClipInt {
        value,
        error: ptr::null_mut(),
    }
}

fn err_int(msg: &str) -> WlClipInt {
    WlClipInt {
        value: -1,
        error: CString::new(msg).unwrap().into_raw(),
    }
}

#[no_mangle]
pub extern "C" fn wlclip_set_foreground(val: bool) {
    FOREGROUND.store(val, Ordering::SeqCst);
}

#[no_mangle]
pub extern "C" fn wlclip_get_text() -> WlClipString {
    let result = get_contents(ClipboardType::Regular, Seat::Unspecified, MimeType::Text);
    match result {
        Ok((mut pipe, _)) => {
            let mut contents = Vec::new();
            match pipe.read_to_end(&mut contents) {
                Ok(_) => {
                    let text = String::from_utf8_lossy(&contents).into_owned();
                    ok_string(text)
                }
                Err(e) => err_string(&format!("read failed: {}", e)),
            }
        }
        Err(ClipboardError::NoSeats)
        | Err(ClipboardError::ClipboardEmpty)
        | Err(ClipboardError::NoMimeType) => err_string("clipboard is empty or unavailable"),
        Err(e) => err_string(&format!("clipboard error: {}", e)),
    }
}

#[no_mangle]
pub extern "C" fn wlclip_set_text(text: *const c_char) -> WlClipInt {
    if text.is_null() {
        return err_int("text pointer is null");
    }

    let c_str = unsafe { CStr::from_ptr(text) };
    let text_str = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => return err_int("failed to convert C string to UTF-8"),
    };

    let mut opts = CopyOptions::new();
    opts.foreground(get_foreground());
    match opts.copy(
        Source::Bytes(text_str.as_bytes().into()),
        CopyMimeType::Autodetect,
    ) {
        Ok(()) => ok_int(0),
        Err(e) => err_int(&format!("failed to set clipboard text: {}", e)),
    }
}

#[no_mangle]
pub extern "C" fn wlclip_get_image() -> WlClipBytes {
    let mime_types = [
        "image/png",
        "image/jpeg",
        "image/gif",
        "image/webp",
        "image/avif",
        "image/jxl",
        "image/bmp",
        "image/tiff",
    ];

    for mime in mime_types.iter() {
        let result = get_contents(
            ClipboardType::Regular,
            Seat::Unspecified,
            MimeType::Specific(*mime),
        );
        match result {
            Ok((mut pipe, _)) => {
                let mut contents = Vec::new();
                match pipe.read_to_end(&mut contents) {
                    Ok(_) => return ok_bytes(contents),
                    Err(_) => continue,
                }
            }
            Err(_) => continue,
        }
    }

    err_bytes("clipboard is empty or contains no supported image")
}

#[no_mangle]
pub extern "C" fn wlclip_set_image_type(
    image_data: *const u8,
    len: usize,
    mime_type: *const c_char,
) -> WlClipInt {
    if image_data.is_null() || len == 0 {
        return err_int("image data pointer is null or length is zero");
    }

    if mime_type.is_null() {
        return err_int("mime type is null");
    }

    let c_str = unsafe { CStr::from_ptr(mime_type) };
    let mime = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => return err_int("failed to convert mime type to UTF-8"),
    };

    let bytes = unsafe { std::slice::from_raw_parts(image_data, len) };

    let mut opts = CopyOptions::new();
    opts.foreground(get_foreground());
    match opts.copy(
        Source::Bytes(bytes.into()),
        CopyMimeType::Specific(mime.to_string()),
    ) {
        Ok(()) => ok_int(0),
        Err(e) => err_int(&format!("failed to set clipboard image: {}", e)),
    }
}

#[no_mangle]
pub extern "C" fn wlclip_get_files() -> WlClipString {
    let result = get_contents(
        ClipboardType::Regular,
        Seat::Unspecified,
        MimeType::Specific("text/uri-list"),
    );
    match result {
        Ok((mut pipe, _)) => {
            let mut contents = Vec::new();
            match pipe.read_to_end(&mut contents) {
                Ok(_) => {
                    let uri_list = String::from_utf8_lossy(&contents);
                    let files: Vec<String> = uri_list
                        .lines()
                        .filter(|line| line.starts_with("file://"))
                        .map(|line| {
                            let path = line.trim_start_matches("file://");
                            let decoded = if path.starts_with("//") {
                                percent_decode_str(&path[1..])
                            } else {
                                percent_decode_str(path)
                            };
                            decoded
                                .decode_utf8()
                                .map(|s| s.into_owned())
                                .unwrap_or_else(|_| path.to_string())
                        })
                        .collect();

                    match serde_json::to_string(&files) {
                        Ok(json) => ok_string(json),
                        Err(e) => err_string(&format!("failed to serialize files: {}", e)),
                    }
                }
                Err(e) => err_string(&format!("failed to read file list: {}", e)),
            }
        }
        Err(ClipboardError::NoSeats)
        | Err(ClipboardError::ClipboardEmpty)
        | Err(ClipboardError::NoMimeType) => err_string("clipboard is empty or contains no files"),
        Err(e) => err_string(&format!("clipboard error: {}", e)),
    }
}

#[no_mangle]
pub extern "C" fn wlclip_set_files(json: *const c_char) -> WlClipInt {
    if json.is_null() {
        return err_int("json pointer is null");
    }

    let c_str = unsafe { CStr::from_ptr(json) };
    let json_str = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => return err_int("failed to convert C string to UTF-8"),
    };

    let files: Vec<String> = match serde_json::from_str(json_str) {
        Ok(f) => f,
        Err(e) => return err_int(&format!("failed to parse files JSON: {}", e)),
    };

    let uri_list: String = files
        .iter()
        .map(|path| format!("file://{}", path))
        .collect::<Vec<_>>()
        .join("\r\n");

    let mut opts = CopyOptions::new();
    opts.foreground(get_foreground());
    match opts.copy(
        Source::Bytes(uri_list.into_bytes().into()),
        CopyMimeType::Specific("text/uri-list".to_string()),
    ) {
        Ok(()) => ok_int(0),
        Err(e) => err_int(&format!("failed to set clipboard files: {}", e)),
    }
}

#[no_mangle]
pub extern "C" fn wlclip_free_string(ptr: *mut c_char) {
    if !ptr.is_null() {
        unsafe { drop(CString::from_raw(ptr)) };
    }
}

#[no_mangle]
pub extern "C" fn wlclip_free_bytes(ptr: *mut u8, len: usize) {
    if !ptr.is_null() {
        unsafe {
            drop(Box::from_raw(std::ptr::slice_from_raw_parts_mut(ptr, len)));
        }
    }
}
