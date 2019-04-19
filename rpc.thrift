exception Error {
    1: string message;
}

service I2C {
    string String(),
    binary Tx(1:i16 addr, 2:binary w, 3:i32 length) throws (1:Error err1),
    void SetSpeed(1:i64 microHertz) throws (1:Error err1),
}