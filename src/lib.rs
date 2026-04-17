use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::sync::{Arc, Mutex};

use numbat::{
    markup::plain_text_format, module_importer::BuiltinModuleImporter, resolver::CodeSource,
    Context, FormatOptions, InterpreterSettings,
};

pub struct NumbatWrapper {
    ctx: Context,
}

#[repr(C)]
pub struct NumbatResult {
    pub out: *mut c_char,
    pub err: *mut c_char,
    pub value: f64,
    pub is_quantity: bool,
    pub unit: *mut c_char,
}

/// Creates a new Numbat context and loads the prelude
#[unsafe(no_mangle)]
pub extern "C" fn numbat_init() -> *mut NumbatWrapper {
    let mut ctx = Context::new(BuiltinModuleImporter::default());
    let _ = ctx.interpret("use prelude", CodeSource::Internal);

    Box::into_raw(Box::new(NumbatWrapper { ctx }))
}

/// Evaluates a block of Numbat code and returns the result or an error
#[unsafe(no_mangle)]
pub extern "C" fn numbat_interpret(
    wrapper: *mut NumbatWrapper,
    code: *const c_char,
) -> NumbatResult {
    let wrapper = unsafe {
        assert!(!wrapper.is_null());
        &mut *wrapper
    };

    let code_str = unsafe {
        assert!(!code.is_null());
        CStr::from_ptr(code).to_str().unwrap()
    };

    // Buffer to capture stdout (e.g. from `print(x)`)
    let printed_output = Arc::new(Mutex::new(String::new()));
    let printed_output_clone = printed_output.clone();

    let mut settings = InterpreterSettings {
        print_fn: Box::new(move |s| {
            let mut out = printed_output_clone.lock().unwrap();
            out.push_str(&plain_text_format(s, false).to_string());
            out.push('\n');
        }),
    };

    match wrapper
        .ctx
        .interpret_with_settings(&mut settings, code_str, CodeSource::Text)
    {
        Ok((statements, result)) => {
            let mut out = printed_output.lock().unwrap().clone();
            let mut val_f64 = 0.0;
            let mut is_q = false;
            let mut unit_str = String::new();

            if result.is_value() {
                let result_markup = result.to_markup(
                    statements.last(),
                    wrapper.ctx.dimension_registry(),
                    false, // don't show type info `[Velocity]`
                    true,  // show leading `= `
                    &FormatOptions::default(),
                );
                out.push_str(&plain_text_format(&result_markup, false).to_string());

                // Extract the raw float value and unit if it's a valid quantity
                if let numbat::InterpreterResult::Value(numbat::value::Value::Quantity(q)) = result
                {
                    val_f64 = q.unsafe_value().to_f64();
                    is_q = true;
                    unit_str = q.unit().to_string();
                }
            }

            NumbatResult {
                out: CString::new(out.trim().to_string()).unwrap().into_raw(),
                err: std::ptr::null_mut(),
                value: val_f64,
                is_quantity: is_q,
                unit: if is_q {
                    CString::new(unit_str).unwrap().into_raw()
                } else {
                    std::ptr::null_mut()
                },
            }
        }
        Err(e) => {
            let error_msg = match &*e {
                numbat::NumbatError::TypeCheckError(
                    numbat::TypeCheckError::IncompatibleDimensions(err),
                ) => {
                    format!("Incompatible dimensions in {}:\n{}", err.operation, err)
                }
                _ => e.to_string(),
            };

            NumbatResult {
                out: std::ptr::null_mut(),
                err: CString::new(error_msg).unwrap().into_raw(),
                value: 0.0,
                is_quantity: false,
                unit: std::ptr::null_mut(),
            }
        }
    }
}

/// Inject a variable directly into the Numbat context
#[unsafe(no_mangle)]
pub extern "C" fn numbat_set_variable(
    wrapper: *mut NumbatWrapper,
    name: *const c_char,
    value: f64,
    unit: *const c_char,
) -> *mut c_char {
    let wrapper = unsafe {
        assert!(!wrapper.is_null());
        &mut *wrapper
    };

    let name_str = unsafe { CStr::from_ptr(name).to_str().unwrap() };
    let unit_str = if unit.is_null() {
        ""
    } else {
        unsafe { CStr::from_ptr(unit).to_str().unwrap() }
    };

    // Construct the Numbat assignment code: `let name = value unit`
    let code = if unit_str.is_empty() {
        format!("let {} = {}", name_str, value)
    } else {
        format!("let {} = {} {}", name_str, value, unit_str)
    };

    match wrapper.ctx.interpret(&code, CodeSource::Internal) {
        Ok(_) => std::ptr::null_mut(), // Success returns null
        Err(e) => CString::new(e.to_string()).unwrap().into_raw(),
    }
}

/// Frees the strings contained within a NumbatResult
#[unsafe(no_mangle)]
pub extern "C" fn numbat_free_result(res: NumbatResult) {
    if !res.out.is_null() {
        unsafe {
            drop(CString::from_raw(res.out));
        }
    }
    if !res.err.is_null() {
        unsafe {
            drop(CString::from_raw(res.err));
        }
    }
    if !res.unit.is_null() {
        unsafe {
            drop(CString::from_raw(res.unit));
        }
    }
}

/// Frees a generic string returned from Rust
#[unsafe(no_mangle)]
pub extern "C" fn numbat_free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            drop(CString::from_raw(s));
        }
    }
}

/// Frees the Numbat context
#[unsafe(no_mangle)]
pub extern "C" fn numbat_free(wrapper: *mut NumbatWrapper) {
    if !wrapper.is_null() {
        unsafe {
            drop(Box::from_raw(wrapper));
        }
    }
}
