/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2021 by Andrew Binstock. All rights reserved.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0)
 */

package classloader

import (
	"errors"
	"jacobin/log"
	"strconv"
	"strings"
)

// Performs the format check on a fully parsed class. The requirements are listed
// here: https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.8
// They are:
// 1) must start with 0xCAFEBABE -- this is verified in the parsing, so not done here
// 2) most predefined attributes must be the right length -- verified during parsing
// 3) class must not be truncated or have extra bytes -- verified during parsing
// 4) CP must fulfill all constraints. This is done in this function
// 5) Fields must have valid names, classes, and descriptions. Partially done in
//    the parsing, but entirely done below
func formatCheckClass(klass *parsedClass) error {
	err := validateConstantPool(klass)
	if err != nil {
		return err // whatever error occurs, the user will have been notified
	}

	err = validateFields(klass)
	return err
}

// validates that the CP fits all the requirements enumerated in:
// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4
// some of these checks were performed perforce in the parsing. Here, however,
// we verify them all. This is a requirement of all classes loaded in the JVM
// Note that this is *not* part of the larger class verification process.
func validateConstantPool(klass *parsedClass) error {
	cpSize := klass.cpCount
	if len(klass.cpIndex) != cpSize {
		return cfe("Error in size of constant pool discovered in format check." +
			"Expected: " + strconv.Itoa(cpSize) + ", got: " + strconv.Itoa(len(klass.cpIndex)))
	}

	if klass.cpIndex[0].entryType != Dummy {
		return cfe("Missing dummy entry in first slot of constant pool")
	}

	for j := 1; j < cpSize; j++ {
		entry := klass.cpIndex[j]
		switch entry.entryType {
		case UTF8:
			// points to an entry in utf8Refs, which holds a string. Check for:
			// * No byte may have the value (byte)0.
			// * No byte may lie in the range (byte)0xf0 to (byte)0xff
			whichUtf8 := entry.slot
			if whichUtf8 < 0 || whichUtf8 >= len(klass.utf8Refs) {
				return cfe("CP entry #" + strconv.Itoa(j) + "points to invalid UTF8 entry: " +
					strconv.Itoa(whichUtf8))
			}
			utf8string := klass.utf8Refs[whichUtf8].content
			utf8bytes := []byte(utf8string)
			for _, char := range utf8bytes {
				if char == 0x00 || (char >= 0xf0 && char <= 0xff) {
					return cfe("UTF8 string for CP entry #" + strconv.Itoa(j) +
						" contains an invalid character")
				}
			}
		case IntConst:
			// there are no specific format checks for integers, so we only check
			// that there is a valid entry pointed to in intConsts
			whichInt := entry.slot
			if whichInt < 0 || whichInt >= len(klass.intConsts) {
				return cfe("Integer at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP intConsts")
			}
		case FloatConst:
			// there are complex bit patterns that can be enforced for floats, but
			// for the nonce, we'll just make sure that the float index points to an actual value
			whichFloat := entry.slot
			if whichFloat < 0 || whichFloat >= len(klass.floats) {
				return cfe("Float at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP floats")
			}
		case LongConst:
			// there are complex bit patterns that can be enforced for longs, but for the
			// nonce, we'll just make sure that there is an actual value pointed to and
			// that the long is followed in the CP by a dummy entry. Consult:
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.5
			whichLong := entry.slot
			if whichLong < 0 || whichLong >= len(klass.longConsts) {
				return cfe("Long constant at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP longConsts")
			}

			nextEntry := klass.cpIndex[j+1]
			if nextEntry.entryType != Dummy {
				return cfe("Missing dummy entry after long constant at CP entry#" +
					strconv.Itoa(j))
			}
			j += 1
		case DoubleConst:
			// see the comments on the LongConst. They apply exactly to the following code.
			whichDouble := entry.slot
			if whichDouble < 0 || whichDouble >= len(klass.doubles) {
				return cfe("Double constant at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP doubless")
			}

			nextEntry := klass.cpIndex[j+1]
			if nextEntry.entryType != Dummy {
				return cfe("Missing dummy entry after double constant at CP entry#" +
					strconv.Itoa(j))
			}
			j += 1
		case ClassRef:
			// the only field of a ClassRef points to a UTF8 entry holding the class name
			// in the case of arrays, the UTF8 entry will describe the type and dimensions of the array
			whichClassRef := entry.slot
			if whichClassRef < 0 || whichClassRef >= len(klass.utf8Refs) {
				return cfe("ClassRef at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP utf8Refs")
			}
		case StringConst:
			// a StringConst holds only an index into the utf8Refs. so we check this.
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.3
			whichString := entry.slot
			if whichString < 0 || whichString >= len(klass.utf8Refs) {
				return cfe("Constant String at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP utf8Refs")
			}
		case FieldRef:
			// the requirements are that the class index points to a valid Class entry
			// and the name_and_type index points to a valid NameAndType entry. Consult
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.2
			// Here we just make sure they point to entries of the correct type and that
			// they exist. The pointed-to entries are themselves validated as this loop
			// picks them up going through the CP.
			whichFieldRef := entry.slot
			if whichFieldRef < 0 || whichFieldRef >= len(klass.fieldRefs) {
				return cfe("Field Ref at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP fieldRefs")
			}
			fieldRef := klass.fieldRefs[whichFieldRef]
			classIndex := fieldRef.classIndex
			class := klass.cpIndex[classIndex]
			if class.entryType != ClassRef ||
				class.slot < 0 || class.slot >= len(klass.classRefs) {
				return cfe("Field Ref at CP entry #" + strconv.Itoa(j) +
					" has a class index that points to an invalid entry in ClassRefs. " +
					strconv.Itoa(classIndex))
			}

			nameAndType := klass.cpIndex[fieldRef.nameAndTypeIndex]
			if nameAndType.entryType != NameAndType ||
				nameAndType.slot < 0 || nameAndType.slot >= len(klass.nameAndTypes) {
				return cfe("Field Ref at CP entry #" + strconv.Itoa(j) +
					" has a nameAndType index that points to an invalid entry in nameAndTypes. " +
					strconv.Itoa(fieldRef.nameAndTypeIndex))
			}
		case MethodRef:
			// the MethodRef must have a class index that points to a Class_info entry
			// which itself must point to a class, not an interface. The MethodRef also has
			// an index to a NameAndType entry. If the name of the latter entry begins with
			// and <, then the name can only be <init>. Consult:
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.2
			whichMethodRef := entry.slot
			methodRef := klass.methodRefs[whichMethodRef]

			classIndex := methodRef.classIndex
			class := klass.cpIndex[classIndex]
			if class.entryType != ClassRef ||
				class.slot < 0 || class.slot >= len(klass.classRefs) {
				return cfe("Method Ref at CP entry #" + strconv.Itoa(j) +
					" holds an invalid class index: " +
					strconv.Itoa(class.slot))
			}

			nAndTIndex := methodRef.nameAndTypeIndex
			nAndT := klass.cpIndex[nAndTIndex]
			if nAndT.entryType != NameAndType ||
				nAndT.slot < 0 || nAndT.slot >= len(klass.nameAndTypes) {
				return cfe("Method Ref at CP entry #" + strconv.Itoa(j) +
					" holds an invalid NameAndType index: " +
					strconv.Itoa(nAndT.slot))
			}

			nAndTentry := klass.nameAndTypes[nAndT.slot]
			methodNameIndex := nAndTentry.nameIndex
			name, err := fetchUTF8string(klass, methodNameIndex)
			if err != nil {
				return cfe("Method Ref (at CP entry #" + strconv.Itoa(j) +
					") has a Name and Type entry does not have a name that is a valid UTF8 entry")
			}

			nameBytes := []byte(name)
			if nameBytes[0] == '<' && name != "<init>" {
				return cfe("Method Ref at CP entry #" + strconv.Itoa(j) +
					" holds an NameAndType index to an entry with an invalid method name " +
					name)
			}
		case Interface:
			// the Interface entries are almost identical to the class entries (see above),
			// except that the class index must point to an interface class, and the requirement
			// re naming < and <init> does not apply.
			whichInterface := entry.slot
			interfaceRef := klass.interfaceRefs[whichInterface]

			classIndex := interfaceRef.classIndex
			class := klass.cpIndex[classIndex]
			if class.entryType != ClassRef ||
				class.slot < 0 || class.slot >= len(klass.classRefs) {
				return cfe("Interface Ref at CP entry #" + strconv.Itoa(j) +
					" holds an invalid class index: " + strconv.Itoa(class.slot))
			}

			clRef := klass.classRefs[class.slot]
			// utfIndex, err := fetchUTF8slot(klass, clRef)
			_, err := fetchUTF8slot(klass, clRef)
			if err != nil {
				return cfe("Interface Ref at CP entry #" + strconv.Itoa(j) +
					" holds an invalid UTF8 index to the interface name: " +
					strconv.Itoa(clRef))
			}

			/* TO REVISIT: with java.lang.String the following code works OK
			with the three interfaces defined in klass.interfaces[], but Iterable
			is not among those classes and yet it's got a interfaceRef CP entry.
			So, not presently sure how you validate that the interfaceRef CP entry
			points to an interface. So for the nonce, the following code is commented out.

			// now that we have the UTF8 index for the interface reference,
			// check whether it's in our list of interfaces for this class.
			matchesInterface := false
			for i := range klass.interfaces {
				if klass.interfaces[i] == utfIndex {
					matchesInterface = true
				}
			}

			if ! matchesInterface {
				return cfe("Interface Ref at CP entry #"+ strconv.Itoa(j) +
					" does not match to any interface in this class.")
			}
			*/

			nAndTIndex := interfaceRef.nameAndTypeIndex
			nAndT := klass.cpIndex[nAndTIndex]
			if nAndT.entryType != NameAndType ||
				nAndT.slot < 0 || nAndT.slot >= len(klass.nameAndTypes) {
				return cfe("Method Ref at CP entry #" + strconv.Itoa(j) +
					" holds an invalid NameAndType index: " +
					strconv.Itoa(nAndT.slot))
			}
		case NameAndType:
			// a NameAndType entry points to two UTF8 entries: name and description. Consult
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.6
			// the descriptor points either to a method, whose UTF8 should begin with a (
			// or to a field, which must start with one of the letter specified in:
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.3.2-200
			whichNandT := entry.slot
			if whichNandT < 0 || whichNandT >= len(klass.nameAndTypes) {
				return cfe("Name and Type at CP entry #" + strconv.Itoa(j) +
					" points to an invalid entry in CP nameAndTypes")
			}

			nAndTentry := klass.nameAndTypes[whichNandT]
			_, err := fetchUTF8string(klass, nAndTentry.nameIndex)
			if err != nil {
				return cfe("Name and Type at CP entry #" + strconv.Itoa(j) +
					" has a name index that points to an invalid UTF8 entry: " +
					strconv.Itoa(nAndTentry.nameIndex))
			}

			desc, err2 := fetchUTF8string(klass, nAndTentry.descriptorIndex)
			if err2 != nil {
				return cfe("Name and Type at CP entry #" + strconv.Itoa(j) +
					" has a description index that points to an invalid UTF8 entry: " +
					strconv.Itoa(nAndTentry.nameIndex))
			}

			err = validateFieldDesc(desc)
			if err != nil {
				return cfe("Name and Type at CP entry #" + strconv.Itoa(j) +
					" has an invalid description string: " + desc)
			}
		case MethodHandle:
			// Method handles have complex validation logic. It's entirely enforced here. See:
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.8
			// CONSTANT_MethodHandle_info {
			//    u1 tag;
			//    u1 reference_kind;
			//    u2 reference_index; }
			whichMethHandle := entry.slot
			mhe := klass.methodHandles[whichMethHandle]
			refKind := mhe.referenceKind
			if refKind < 1 || refKind > 9 {
				return cfe("MethodHandle at CP entry #" + strconv.Itoa(j) +
					" has an invalid reference kind: " + strconv.Itoa(refKind))
			}
			refIndex := mhe.referenceIndex

			switch refKind {
			// if refKind is 1-4, the reference_index must point to a fieldRef
			case 1, 2, 3, 4:
				if klass.cpIndex[refIndex].entryType != FieldRef {
					return cfe("MethodHandle at CP entry #" + strconv.Itoa(j) +
						" has an reference kind between 1-4 ( " + strconv.Itoa(refKind) +
						") which does not point to a FieldRef")
				}
			// if refKind is 5 or 8, the reference_index must point to a methodRef
			case 5, 8:
				if klass.cpIndex[refIndex].entryType != MethodRef {
					return cfe("MethodHandle at CP entry #" + strconv.Itoa(j) +
						" has an reference kind between of 5 or 8 ( " + strconv.Itoa(refKind) +
						") which does not point to a MethodRef")
				}
			case 6, 7:
				// if refKind is 6 or 7, the reference_index must point to a methodRef or if the
				// class version # is >= 52, it can point to an Interface. To make the logic readable,
				// we test for the positive here, rather than the negative as in the other cases
				if klass.cpIndex[refIndex].entryType == MethodRef ||
					(klass.javaVersion >= 52 && klass.cpIndex[refIndex].entryType == Interface) {
					break
				} else {
					return cfe("MethodHandle at CP entry #" + strconv.Itoa(j) +
						" has an reference kind between of 6 or 7 ( " + strconv.Itoa(refKind) +
						") which does not point to a MethodRef or in Java version 52 or later " +
						"does not point to an Interface.")
				}
			case 9:
				if klass.cpIndex[refIndex].entryType != Interface {
					return cfe("MethodHandle at CP entry #" + strconv.Itoa(j) +
						" has an reference kind between of 9 which does not point to an interface")
				}
			}

			// get the class name pointed to by the MethodRef pointed to by the MethodHandle
			methodName, _, _, err := resolveCPmethodRef(refIndex, klass)
			if err != nil {
				return errors.New("") // the error messsage is already displayed
			}

			// if the reference_kind is 5-7 the name of the method pointed to
			// by the nameAndType entry in the method handle cannot be <init> or <clinit>
			if refKind >= 5 && refKind <= 7 && klass.cpIndex[refIndex].entryType == MethodRef {
				methRefIndex := klass.cpIndex[refIndex].slot
				if methRefIndex < 0 || methRefIndex >= len(klass.methodRefs) {
					return cfe("Reference index for MethodHandle at CP entry #" + strconv.Itoa(j) +
						" points to an invalid MethodRef: " + strconv.Itoa(methRefIndex))
				}

				if methodName == "<init>" || methodName == "<clinit>" {
					return cfe("Invalid class name for MethodHandle at CP entry #" + strconv.Itoa(j) +
						" : " + methodName)
				}
			} else if refKind == 8 {
				if methodName != "<init>" {
					return cfe("Class name for MethodHandle at CP entry #" + strconv.Itoa(j) +
						" should be <init>, but is: " + methodName)
				}
			}

			log.Log("ClassName in MethodRef of MethodHandle at CP entry #"+strconv.Itoa(j)+
				" is:"+methodName, log.FINEST)
		case MethodType:
			// Method types consist of an integer pointing to a CP entry that's a UTF8 description
			// of the method type, which appears to require an initial opening parenthesis. See
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.9
			whichMethType := entry.slot
			mte := klass.methodTypes[whichMethType]
			utf8 := klass.cpIndex[mte]
			if utf8.entryType != UTF8 || utf8.slot < 0 || utf8.slot > len(klass.utf8Refs)-1 {
				return cfe("MethodType at CP entry #" + strconv.Itoa(j) +
					" has an invalid description index: " + strconv.Itoa(utf8.slot))
			}
			methType := klass.utf8Refs[utf8.slot]
			if !strings.HasPrefix(methType.content, "(") {
				return cfe("MethodType at CP entry #" + strconv.Itoa(j) +
					" does not point to a type that starts with an open parenthesis. Got: " +
					methType.content)
			}
		case InvokeDynamic:
			// InvokeDynamic is a unique kind of entry. The first field, boostrapIndex, must be a
			// "valid index into the bootstrap_methods array of the bootstrap method table of this
			// this class file" (specified in §4.7.23). The document spec for InvokeDynamic entries is:
			// https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-4.4.10
			// Once we actually get bootstrap entry table of the method, we'll circle back here to
			// check it. The second field is a nameAndType record describing the boostrap method.
			// Here we just make sure, the field points to the right kind of entry. That entry
			// will be checked later/earlier in this CP checking loop.
			whichInvDyn := entry.slot
			invDyn := klass.invokeDynamics[whichInvDyn]

			// bootstrap = invDyn.bootstrapIndex // TODO: Check the boostrap entry as soon as we can
			nAndT := invDyn.nameAndType
			if nAndT < 1 || nAndT > len(klass.cpIndex)-1 {
				return cfe("The entry number into klass.InvokeDynamics[] at CP entry #" +
					strconv.Itoa(j) + " is invalid: " + strconv.Itoa(nAndT))
			}
			if klass.cpIndex[nAndT].entryType != NameAndType {
				return cfe("NameAndType index at CP entry #" + strconv.Itoa(j) +
					" (InvokeDynamic) points to an entry that's not NameAndType: " +
					strconv.Itoa(klass.cpIndex[nAndT].entryType))
			}

			// TODO: continue format checking other CP entries
		default:
			continue
		}
	}

	return nil
}

// field entries consist of two string indexes, one of which points to the name, the other
// to a string containing a description of the type. Here we grab the strings and check that
// they fulfill the requirements: name doesn't start with a digit or contain a space, and the
// type begins with one of the required letters/symbols
func validateFields(klass *parsedClass) error {
	for i, f := range klass.fields {
		// f.name points to a UTF8 entry in klass.utf8refs, so check it's in a valid range
		if f.name < 0 || f.name >= len(klass.utf8Refs) {
			return cfe("Invalid index to UTF8 string for field name in field #" + strconv.Itoa(i))
		}
		fName := klass.utf8Refs[f.name].content

		// f.description points to a UTF8 entry in klass.utf8refs, so check it's in a valid range
		if f.description < 0 || f.description >= len(klass.utf8Refs) {
			return cfe("Invalid index for UTF8 string containing description of field " + fName)
		}
		fDesc := klass.utf8Refs[f.description].content

		fNameBytes := []byte(fName)
		if fNameBytes[0] >= '0' && fNameBytes[0] <= '9' {
			return cfe("Invalid field name in format check (starts with a digit): " + fName)
		}

		// check that there is no leading, trailing, or embedded whitespace
		for _, c := range fNameBytes {
			switch c {
			case
				'\u0009', // horizontal tab
				'\u000A', // line feed
				'\u000B', // vertical tab
				'\u000C', // form feed
				'\u000D', // carriage return
				'\u0020', // space
				'\u0085', // next line
				'\u00A0': // no-break space
				return cfe("Invalid field name in format check (contains whitespace): " + fName)
			default:
				continue
			}
		}

		if validateFieldDesc(fDesc) != nil {
			return cfe("Field " + fName + " has an invalid description string: " + fDesc)
		}
	}
	return nil
}

// certain descriptions and type strings must start with one of the letters shown here.
// See: https://docs.oracle.com/javase/specs/jvms/se11/html/jvms-4.html#jvms-FieldType
func validateFieldDesc(desc string) error {
	if len(desc) < 1 {
		return errors.New("invalid")
	}

	descBytes := []byte(desc)
	c := descBytes[0]
	if !(c == '(' || c == 'B' || c == 'C' || c == 'D' || c == 'F' ||
		c == 'I' || c == 'J' || c == 'L' || c == 'S' || c == 'Z' ||
		c == '[') {
		return errors.New("invalid")
	}
	return nil
}
