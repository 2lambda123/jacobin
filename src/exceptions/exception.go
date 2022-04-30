/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2022 by Andrew Binstock. All rights reserved.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0)
 */

package exceptions

import (
	"fmt"
	"jacobin/frames"
	"jacobin/globals"
	"jacobin/log"
	"jacobin/thread"
)

// List of Java exceptions (as of Java 17)
const (
	Unknown = iota

	// runtime exceptions
	AnnotationTypeMismatchException
	ArithmeticException
	ArrayStoreException
	BufferOverflowException
	BufferUnderflowException
	CannotRedoException
	CannotUndoException
	CatalogException
	ClassCastException
	ClassNotPreparedException
	CMMException
	CompletionException
	ConcurrentModificationException
	DateTimeException
	DOMException
	DuplicateRequestException
	EmptyStackException
	EnumConstantNotPresentException
	EventException
	FileSystemAlreadyExistsException
	FileSystemNotFoundException
	FindException
	IllegalArgumentException
	IllegalCallerException
	IllegalMonitorStateException
	IllegalPathStateException
	IllegalStateException
	IllformedLocaleException
	ImagingOpException
	InaccessibleObjectException
	IncompleteAnnotationException
	InconsistentDebugInfoException
	IndexOutOfBoundsException
	InternalException
	InvalidCodeIndexException
	InvalidLineNumberException
	InvalidModuleDescriptorException
	InvalidModuleException
	InvalidRequestStateException
	InvalidStackFrameException
	JarSignerException
	JMRuntimeException
	JSException
	LayerInstantiationException
	LSException
	MalformedParameterizedTypeException
	MalformedParametersException
	MirroredTypesException
	MissingResourceException
	NativeMethodException
	NegativeArraySizeException
	NoSuchDynamicMethodException
	NoSuchElementException
	NoSuchMechanismException
	NullPointerException
	ObjectCollectedException
	ProfileDataException
	ProviderException
	ProviderNotFoundException
	RangeException
	RasterFormatException
	RejectedExecutionException
	ResolutionException
	SecurityException
	SPIResolutionException
	TypeNotPresentException
	UncheckedIOException
	UndeclaredThrowableException
	UnknownEntityException
	UnmodifiableModuleException
	UnmodifiableSetException
	UnsupportedOperationException
	VMDisconnectedException
	VMMismatchException
	VMOutOfMemoryException
	WrongMethodTypeException
	XPathException

	// non-runtime exceptions
	AbsentInformationException
	AclNotFoundException
	ActivationException
	AgentInitializationException
	AgentLoadException
	AlreadyBoundException
	AttachNotSupportedException
	AWTException
	BackingStoreException
	BadAttributeValueExpException
	BadBinaryOpValueExpException
	BadLocationException
	BadStringOperationException
	BrokenBarrierException
	CardException
	CertificateException
	ClassNotLoadedException
	CloneNotSupportedException
	DataFormatException
	DatatypeConfigurationException
	DestroyFailedException
	ExecutionControl
	ExecutionControlException
	ExecutionException
	ExpandVetoException
	FontFormatException
	GeneralSecurityException
	GSSException
	IllegalClassFormatException
	IllegalConnectorArgumentsException
	IncompatibleThreadStateException
	InterruptedException
	IntrospectionException
	InvalidApplicationException
	InvalidMidiDataException
	InvalidPreferencesFormatException
	InvalidTargetObjectTypeException
	InvalidTypeException
	InvocationException
	IOException
	JMException
	JShellException
	KeySelectorException
	LambdaConversionException
	LastOwnerException
	LineUnavailableException
	MarshalException
	MidiUnavailableException
	MimeTypeParseException
	NamingException
	NoninvertibleTransformException
	NotBoundException
	NotOwnerException
	ParseException
	ParserConfigurationException
	PrinterException
	PrintException
	PrivilegedActionException
	PropertyVetoException
	ReflectiveOperationException
	RefreshFailedException
	RuntimeException
	SAXException
	ScriptException
	ServerNotActiveException
	SQLException
	StringConcatException
	TimeoutException
	TooManyListenersException
	TransformerException
	TransformException
	UnmodifiableClassException
	UnsupportedAudioFileException
	UnsupportedCallbackException
	UnsupportedFlavorException
	UnsupportedLookAndFeelException
	URIReferenceException
	URISyntaxException
	VMStartException
	XAException
	XMLParseException
	XMLSignatureException
	XMLStreamException

	// Java errors
	AnnotationFormatError
	AssertionError
	AWTError
	CoderMalfunctionError
	FactoryConfigurationError
	IOError
	LinkageError
	SchemaFactoryConfigurationError
	ServiceConfigurationError
	ThreadDeath
	TransformerFactoryConfigurationError
	VirtualMachineError
)

var JacobinRuntimeErrLiterals = []string{
	"",
	"",
	"Arithmetic Exception, Divide by Zero",
}

var JDKRuntimeErrLiterals = []string{
	"",
	"",
	"java.lang.ArithmeticException: / by zero",
}

// Throw duplicates the exception mechanism in Java. Right now, it displays the
// error message. Will add: catch logic, stack trace, and halt of execution
// TODO: use ThreadNum to find the right thread
func Throw(excType int, clName string, threadNum int, methName string, cp int) {
	thd := globals.GetGlobalRef().Threads.ThreadsList.Front().Value.(*thread.ExecThread)
	frameStack := thd.Stack
	f := frames.PeekFrame(frameStack, 0)
	fmt.Println("class name: " + f.ClName)
	msg := fmt.Sprintf(
		"%s%sin %s, in%s, at bytecode[]: %d", JacobinRuntimeErrLiterals[excType], ": ", clName, methName, cp)
	_ = log.Log(msg, log.SEVERE)
}

// JVMexception reports runtime exceptions occurring in the JVM (rather than in the app)
// such as invalid JAR files, and the like. For the moment, it prints out the error msg
// only. Eventually, it will print out considerably more info depending on the setting of
// globals.JVMstrict
func JVMexception(excType int, msg string) {
	_ = log.Log(msg, log.SEVERE)
}
