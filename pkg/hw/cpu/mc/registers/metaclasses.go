package registers

// Register metaclass for all integer sized integer registers
var IntegerRegisters *RegisterMetaClass = MakeRegisterMetaClass("IntegerRegisters", []RegisterClass{RegisterClass_GeneralPurposeInteger, RegisterClass_StateRegisters})
