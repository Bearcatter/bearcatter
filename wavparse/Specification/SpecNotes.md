# Uniden WAV File Specification

## Others work

### deim
[From deim's Windows app Uniden Wav Player](https://forums.radioreference.com/threads/uniden-wav-files-player-and-organizer.393944/)

Have requested spec details, no response yet.

In this folder named `deim.zip`.

CSV used for testing was generated from this application.

# gariac

[From gariac's bash script](https://forums.radioreference.com/threads/free-software-bcdx36hp-wav-header-reader.286454/#post-2232183)

Doesn't work well but could be useful.

In this folder named `gariac.sh`.

### theaton
[From theaton's Excel spreadsheet tool](https://forums.radioreference.com/threads/wav-file-manager.287928/)

In this folder named `theaton - BCD436HP WAV File Manager.xls`.

#### Spec

| Pos | Field List & Descriptions                 |
|-----|-------------------------------------------|
| 33  | IART - System name                        |
| 105 | IGNR - Department name                    |
| 177 | INAM - Channel name                       |
| 249 | ICMT - TGID (conventional P25)            |
| 261 | ICMT - Mystery number                     |
| 321 | IPRD - Scanner name                       |
| 347 | IKEY - Related to system name             |
| 377 | ICRD - Closing date/time                  |
| 401 | ISRC - Tone or NAC                        |
| 425 | ITCH - Unit ID                            |
| 497 | ISBJ - Favorite List name                 |
| 569 | ICOP - **************** (skipped)         |
| 593 | Favorite List block (begins with FL name) |
| 658 | System block (begins with System name)    |
| 723 | Department block (begins with Dept. name) |
| 788 | Channel block (begins with Channel name)  |
| 853 | Site block (begins with Site name)        |
| 918 | TGID (trunked only)                       |
| 940 | WACN and RFSS (not yet implemented)       |


| Pos | Favorites List Block   | Truncated? |
|-----|------------------------|------------|
| 1   | Favorite List name     |            |
| 2   | Favorite List filename |            |
| 3   | Location control       |            |
| 4   | Monitor                |            |
| 5   | FL Quick Key           |            |
| 6   | FL Number Tag          |            |
| 7   | Config Key 0           |            |
| 8   | Config Key 1           |            |
| 9   | Config Key 2           |            |
| 10  | Config Key 3           |            |
| 11  | Config Key 4           |            |
| 12  | Config Key 5           |            |
| 13  | Config Key 6           |            |
| 14  | Config Key 7           | Yes        |
| 15  | Config Key 8           | Yes        |
| 16  | Config Key 9           | Yes        |

| Pos | System Block                                |
|-----|---------------------------------------------|
| 1   | System Name                                 |
| 2   | Avoid                                       |
| 3   | Blank                                       |
| 4   | System type (P25 or Ltr)                    |
| 5   | ID Search                                   |
| 6   | Emergency alert type                        |
| 7   | Alert volume                                |
| 8   | Motorola Status Bit                         |
| 9   | P25 NAC (Srch or NAC value)                 |
| 10  | Quick key                                   |
| 11  | Number tag                                  |
| 12  | Hold time                                   |
| 13  | Analog AGC                                  |
| 14  | Digital AGC                                 |
| 15  | End code                                    |
| 16  | Priority ID                                 |
| 17  | Emergency alert light (color)               |
| 18  | Emergency alert condition (fast blink etc.) |

| Pos | Department Block                 |
|-----|----------------------------------|
| 1   | Department Name                  |
| 2   | Avoid                            |
| 3   | Latitude                         |
| 4   | Longitude                        |
| 5   | Range                            |
| 6   | Shape (circle)                   |
| 7   | Department number tag (5 or Off) |
| 8   | Repeated material                |

| Pos | Channel Block                         |
|-----|---------------------------------------|
| 1   | Channel Name                          |
| 2   | Avoid                                 |
| 3   | TGID or Frequency                     |
| 4   | Mode (NFM or DIGITAL)                 |
| 5   | Tone code or settings                 |
| 6   | Service type (converted to numerical) |

| Pos | Channel Block (Conventional) | Truncated? |
|-----|------------------------------|------------|
| 7   | Attenuator                   | Yes        |
| 8   | Delay value                  | Yes        |
| 9   | Volume offset                | Yes        |
| 10  | Alert tone type              | Yes        |
| 11  | Alert tone volume            | Yes        |
| 12  | Alert light color            | Yes        |
| 13  | Alert light type             | Yes        |
| 14  | Number tag                   | Yes        |
| 15  | Priority on/off              | Yes        |

| Pos | Channel Block (Trunked) | Truncated? |
|-----|-------------------------|------------|
| 7   | Delay value             | Yes        |
| 8   | Volume offset           | Yes        |
| 9   | Alert tone type         | Yes        |
| 10  | Alert tone volume       | Yes        |
| 11  | Alert light color       | Yes        |
| 12  | Alert light type        | Yes        |
| 13  | Number tag              | Yes        |
| 14  | Priority on/off         | Yes        |

| Pos | Site Block (Trunked only) | Truncated? |
|-----|---------------------------|------------|
| 1   | Site Name                 |            |
| 2   | Avoid                     |            |
| 3   | Latitude                  |            |
| 4   | Longitude                 |            |
| 5   | Range                     |            |
| 6   | Modulation                | Yes        |
| 7   | Motorola band plan        | Yes        |
| 8   | EDACS (Wide)              | Yes        |
| 9   | Shape (circle)            | Yes        |
| 10  | Attenuator? (On/Off)      | Yes        |

#### Notes
This program extracts WAV file header data from lists of BCD436HP/BCD536HP audio recordings.
The WAV files must have the original filenames and be in the directories created by the scanner.
Copy the folders of WAV files onto a computer in a separate folder with no other files.
Press the "Import Files" button and select the first folder and the first WAV file that you want included.
This program uses home-made directory dialog box to avoid having to ship a DLL file.
This program will process all WAV files and folders with dates later than the file selected.
(Hint: for testing, try a small number of files by starting with a recent one. Large lots can take time.)
A new worksheet will be created with each process (Sheet1, Sheet2, etc.).
These sheets only include the header fields that the developer considered useful and consistent.
Three other sheets (Full, Full>, and Fields) will also be filled with data (and overwritten with each run).
◙ Full - Prints the raw header data with 100 characters per cell, odd characters are converted to "|"
◙ Full> - Same as Full except odd characters are replaced with the ASCII code in brackets "<>"
◙ Fields - Includes all extracted fields, though many are truncated and empty or full of gibberish

If you get a code error, go to the Developer tab and open Visual Basic.
On the Tools tab click References, and make sure the following references are checked:
◙ Visual Basic for Applications
◙ Microsoft Excel 15.0 Object Library
◙ OLE Automation
◙ Microsoft Office 15.0 Object Library
◙ Microsoft Forms 2.0 Object Library
Please inform the developer if you needed to add any of these, or if you have problems or comments.
Please also inform the developer if you discover or refine the identity of any data fields.

This is freeware and a beta version, so there are no guarantees.

© 2014 Timothy H. Heaton, theaton@usd.edu, version 0.3

The BCD436HP/BCD536HP scanners record data in the WAV file headers that can be used for logging.
The standard header blocks (IART, AGNR, etc.) contain data that is not particularly useful.
(Data in these fields end with ASCII-0, with the rest of the block carried over from previous recordings.)
Most of the data is recorded in blocks at the end of the standard header blocks.
Within each large block (FL, System, etc.) the data is delimited with one ASCII-0 character.
When the first field (name) is long, some of the later fields are truncated (common ones noted with "T").

Future plans are to rename the WAV files and sort them into groups.
Renaming could involve pre-appending the TG or Frequency and post-appending the UID or Tone/code.
Sorting could involve grouping in folders by System and further by Department or Site.
Conventional recordings could be grouped either by Department or simply by Frequency.
Your suggestions are welcome.

#### Code

I extracted the Macros so that those without Excel (like me until 10 minutes before writing this) wouldn't have to install it.

##### Parser

```vba
Option Explicit
Public Col%, Row&, MyDir$, User$, Root$, i%, j%, f, fs
'© 2014 Timothy H. Heaton, theaton@usd.edu, version 0.3
Function DateIn(DateAsText$) As Date
  If DateAsText Like "####-##-##_##-##-##*" Then
    DateIn = DateSerial(Left(DateAsText, 4), Mid(DateAsText, 6, 2), Mid(DateAsText, 9, 2)) _
      + TimeSerial(Mid(DateAsText, 12, 2), Mid(DateAsText, 15, 2), Mid(DateAsText, 18, 2))
  ElseIf DateAsText Like "##############" Then
    DateIn = DateSerial(Left(DateAsText, 4), Mid(DateAsText, 5, 2), Mid(DateAsText, 7, 2)) _
      + TimeSerial(Mid(DateAsText, 9, 2), Mid(DateAsText, 11, 2), Mid(DateAsText, 13, 2))
  Else
    MsgBox "File or directory not in proper format." & vbCrLf & DateAsText, , "Error"
    DateIn = 0
  End If
End Function
Sub Import()
  ImportForm.Show 'Open internal Common Dialog Box -- returns only filename (not directory) if in directory from previous selection
  If Root = "" Then Exit Sub
  Sheets.Add After:=Sheets(Sheets.Count)
  Cells(1, 1) = "Please Wait ..."
  Application.ScreenUpdating = False
  ''''''''''''''''''''''''''''''''''''''' Start Test Sheets '''''''''''''''''''''''''''''''''''''''
  Dim aFL() As String, aSys() As String, BigStart, str$, Test As Byte
  Test = 2 '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< 0 for no tests, 1 for Log test, 2 for Full test
  If Test > 0 Then
    If Test > 1 Then
      Sheets("Full").Cells.ClearContents
      Sheets("Full>").Cells.ClearContents
      i = 0
'      Do
'        Sheets("Main").Cells(i + 1, 1) = 0 'Set up columns for Character Count
'        Sheets("Main").Cells(i + 1, 2) = i
'        Sheets("Main").Cells(i + 1, 3) = Hex(i)
'        If i <> 10 Then Sheets("Main").Cells(i + 1, 4) = Chr(i)
'        i = i + 1
'      Loop While i < 256
    End If
    Sheets("Fields").Cells.ClearContents
    ''            IART    IGNR      INAM      ICMT           IPRD      IKEY      ICRD      ISCR      ITCH      ISBJ      ICOP      unid
    'start = Array(25, 33, 97, 105, 169, 177, 241, 254, 261, 313, 321, 337, 345, 369, 377, 393, 402, 417, 425, 489, 497, 561, 569, 585, 593)
  End If
  '''''''''''''''''''''''''''''''''''''''' End Test Sheets ''''''''''''''''''''''''''''''''''''''''
  Dim aByte(980) As Byte, Fil1 As Date, Dir1 As Date, MyFile$, aFile() As String, aDir() As String, iFile%, iDir%
  Dim LocDir$, aStart, aText(17) As String, aDept() As String, aChan() As String, aSite() As String
  aStart = Array(32, 104, 176, 248, 260, 320, 346, 376, 400, 424, 496, 592, 657, 722, 787, 852, 917, 939)
  Cells.NumberFormat = "@"
  Columns("A:A").NumberFormat = "mm/dd/yy hh:mm:ss;@"
  Columns("B:B").NumberFormat = "0"
  Columns("J:J").NumberFormat = "0.0000"
  Columns("M:N").NumberFormat = "0.000000"
  Columns("O:O").NumberFormat = "0"
  Columns("P:Q").NumberFormat = "0.000000"
  Columns("R:R").NumberFormat = "0"
  Rows("1:1").HorizontalAlignment = xlCenter
  Rows("1:1").Font.Bold = True
  Range("B2").Select
  ActiveWindow.FreezePanes = True
  Range("A1").Select
  Cells(1, 1) = "StartTime"
  Cells(1, 2) = "Sec"
  Cells(1, 2).AddComment
  Cells(1, 2).Comment.Text "Duration in Seconds"
  Cells(1, 2).Comment.Shape.TextFrame.AutoSize = True
  Cells(1, 3) = "Favorites List"
  Cells(1, 4) = "System"
  Cells(1, 5) = "Site"
  Cells(1, 6) = "Department"
  Cells(1, 7) = "Channel"
  Cells(1, 8) = "TGID"
  Cells(1, 9) = "UID"
  Cells(1, 10) = "Frequency"
  Cells(1, 11) = "Tone"
  Cells(1, 12) = "Mode"
  Cells(1, 13) = "Site-Lat"
  Cells(1, 14) = "Site-Long"
  Cells(1, 15) = "Rng"
  Cells(1, 16) = "Dept-Lat"
  Cells(1, 17) = "Dept-Long"
  Cells(1, 18) = "Rng"
  f = FreeFile
  Row = 1
  Fil1 = DateIn(Right(Root, 23))
  Dir1 = DateIn(Right(Root, 43))
  Root = Left(Root, Len(Root) - 43)
  LocDir = Dir(Root & "*.*", vbDirectory) 'Reads first file in directory (Error on empty drive)
  ReDim aDir(0)
  Do Until Len(LocDir) = 0 'Get list of directories
    If Len(LocDir) < 3 Then
    ElseIf DateIn(LocDir) >= Dir1 Then
      iDir = UBound(aDir)
      Do Until iDir < 1
        If DateIn(aDir(iDir - 1)) < DateIn(LocDir) Then Exit Do
        aDir(iDir) = aDir(iDir - 1)
        iDir = iDir - 1
      Loop
      aDir(iDir) = LocDir
      ReDim Preserve aDir(UBound(aDir) + 1)
    End If
    LocDir = Dir()
  Loop
  iDir = 0
  Do While iDir < UBound(aDir) 'Loop through directories
    ReDim aFile(0)
    MyFile = Dir(Root & aDir(iDir) & "\*.*", vbDirectory) 'Reads first file in directory (Error on empty drive)
    Do Until Len(MyFile) = 0 'Get list of files
      If Len(MyFile) < 3 Then
      ElseIf DateIn(MyFile) >= Fil1 Then
        iFile = UBound(aFile)
        Do Until iFile < 1
          If DateIn(aFile(iFile - 1)) < DateIn(MyFile) Then Exit Do
          aFile(iFile) = aFile(iFile - 1)
          iFile = iFile - 1
        Loop
        aFile(iFile) = MyFile
        ReDim Preserve aFile(UBound(aFile) + 1)
      End If
      MyFile = Dir()
    Loop
    iFile = 0
    Do While iFile < UBound(aFile) 'Loop through files
      Row = Row + 1
      Open Root & aDir(iDir) & "\" & aFile(iFile) For Binary Access Read As f
      Get f, , aByte()
      Close f
      j = 0
      Col = 0
      Do
        i = aStart(j) 'Extract fixed-length fields
        aText(j) = ""
        Do Until aByte(i) = 0 Or i > UBound(aByte)
          aText(j) = aText(j) & Chr(aByte(i))
          i = i + 1
        Loop
        j = j + 1
      Loop While j < UBound(aStart) + 1
  ''''''''''''''''''''''''''''''''''''''' Start Test Sheets '''''''''''''''''''''''''''''''''''''''
      j = 0
      i = 592 'Extract Favorites List fields
      ReDim aFL(15) '0 useful, 15 max
      Do Until i > 656 Or j > UBound(aFL)
        If aByte(i) = 0 Then
          j = j + 1
        Else
          aFL(j) = aFL(j) & Chr(aByte(i))
        End If
        i = i + 1
      Loop
      j = 0
      i = 657 'Extract System fields
      ReDim aSys(12) '0 useful, 12 max
      Do Until i > 721 Or j > UBound(aSys)
        If aByte(i) = 0 Then
          j = j + 1
        Else
          aSys(j) = aSys(j) & Chr(aByte(i))
        End If
        i = i + 1
      Loop
  '''''''''''''''''''''''''''''''''''''''' End Test Sheets ''''''''''''''''''''''''''''''''''''''''
      j = 0
      i = 722 'Extract Department fields
      ReDim aDept(6) '4 useful, 6 max
      Do Until i > 786 Or j > UBound(aDept)
        If aByte(i) = 0 Then
          j = j + 1
        Else
          aDept(j) = aDept(j) & Chr(aByte(i))
        End If
        i = i + 1
      Loop
      j = 0
      i = 787 'Extract Channel fields
      ReDim aChan(12) '4 useful, 12 max
      Do Until i > 851 Or j > UBound(aChan)
        If aByte(i) = 0 Then
          j = j + 1
        Else
          aChan(j) = aChan(j) & Chr(aByte(i))
        End If
        i = i + 1
      Loop
      j = 0
      i = 852 'Extract Site fields
      ReDim aSite(15) '9 useful, 15 max
      Do Until i > 916 Or j > UBound(aSite)
        If aByte(i) = 0 Then
          j = j + 1
        Else
          aSite(j) = aSite(j) & Chr(aByte(i))
        End If
        i = i + 1
      Loop
      Cells(Row, 1) = DateIn(aFile(iFile)) 'Start Date/Time (A)
      Cells(Row, 2) = Round((DateIn(aText(7)) - Cells(Row, 1)) * 24 * 3600, 0) 'Duration in seconds (B)
      Cells(Row, 3) = aText(11) 'Favorites List (C)
      Cells(Row, 4) = aText(12) 'System Name (D)
      Cells(Row, 6) = aDept(0) 'Department Name (F)
      If aChan(3) <> "ALL" And Not aChan(3) Like "*#*" Then
        Cells(Row, 12) = aChan(3) 'Mode (L)
        If Cells(Row, 12) = "N" Or Cells(Row, 12) = "NF" Then
          Cells(Row, 12) = "NFM"
        ElseIf Cells(Row, 12) = "F" Then
          Cells(Row, 12) = "FM"
        End If
      End If
      If aSite(1) = "Off" Or Len(aSite(0)) > 59 Then 'Trunked recording
        Cells(Row, 5) = aSite(0) 'Site Name (E)
        Cells(Row, 7) = aChan(0) 'Channel Name (G)
        If Len(aText(16)) > 5 Then Cells(Row, 8) = Right(aText(16), Len(aText(16)) - 5) 'Talkgroup ID (H)
        If Len(aText(9)) > 0 Then
          If aText(9) Like "UID:#*" Then
            Cells(Row, 9) = Right(aText(9), Len(aText(9)) - 4) 'Unit ID (I)
          Else
            Cells(Row, 9) = aText(9)
          End If
        End If
        If aSite(2) <> "0.000000" Then
          If aSite(2) Like "*#.######" Then Cells(Row, 13) = aSite(2) 'Site Latitude (M)
          If aSite(3) Like "*#.######" Then Cells(Row, 14) = aSite(3) 'Site Longitude (N)
          If aSite(4) Like "*#.*" Then Cells(Row, 15) = aSite(4) 'Site Range (O)
        End If
      Else 'Conventional recording
        If aChan(0) Like " ###.####MHz" Then
          Cells(Row, 10) = Mid(aChan(0), 2, 8) 'Frequency (J)
        ElseIf aChan(0) Like "  ##.####MHz" Then
          Cells(Row, 10) = Mid(aChan(0), 3, 7)
        ElseIf aChan(0) Like "####.####MHz" Then
          Cells(Row, 10) = Left(aChan(0), 9)
        ElseIf aChan(2) Like "*######00" Then 'Could be in error if truncated
          Cells(Row, 10) = Left(aChan(2), Len(aChan(2)) - 6) & "." & Mid(aChan(2), Len(aChan(2)) - 5, 4)
          Cells(Row, 7) = aChan(0) 'Channel Name (G)
        Else
          Cells(Row, 7) = aChan(0) 'No frequency available
        End If
        If Right(aText(8), 1) = "h" Then
          Cells(Row, 11) = "N" & Mid(aText(8), Len(aText(8)) - 3, 3) 'Tone (K)
        ElseIf aText(8) Like "DCS*###" Then
          Cells(Row, 11) = "D" & Right(aText(8), 3)
        ElseIf aChan(4) Like "TONE=[CD]##[.0-9]*" Then
          Cells(Row, 11) = Right(aChan(4), Len(aChan(4)) - 5) 'Conventional channel format
        ElseIf aChan(4) Like "NAC=[0-9A-F]*" Then
          Cells(Row, 11) = "N" & Format(Right(aChan(4), Len(aChan(4)) - 4), "00#")
        ElseIf aText(8) Like "* ##.#Hz" Then
          Cells(Row, 11) = "C" & Mid(aText(8), Len(aText(8)) - 5, 4) 'Quick Search format
        ElseIf aText(8) Like "*###.#Hz" Then
          Cells(Row, 11) = "C" & Mid(aText(8), Len(aText(8)) - 6, 5)
        ElseIf Len(aText(8)) > 0 Then
          Cells(Row, 11) = aText(8)
        End If
      End If
      If aDept(2) <> "0.000000" Then
        If aDept(2) Like "*#.######" Then Cells(Row, 16) = aDept(2) 'Department Latitude (P)
        If aDept(3) Like "*#.######" Then Cells(Row, 17) = aDept(3) 'Department Longitude (Q)
        If aDept(4) Like "*#.*" Then Cells(Row, 18) = aDept(4) 'Department Range (R)
      End If
      ''''''''''''''''''''''''''''''''''''''' Start Test Sheets '''''''''''''''''''''''''''''''''''''''
      If Test > 0 Then
        If Test > 1 Then
          Col = 0
          str = ""
          Do
            If aByte(Col) < 32 Then
              str = str & "|"
            Else
              str = str & Chr(aByte(Col))
            End If
            Col = Col + 1
          Loop Until Col = 980
          Col = 1
          Do
            Sheets("Full").Cells(Row, Col) = Mid(str, Col * 100 - 99, 100)
            Col = Col + 1
          Loop Until Col > Int(Len(str) / 100)
          Sheets("Full").Cells(Row, Col) = Right(str, Len(str) - (Col - 1) * 100)
          Col = 0
          str = ""
          Do
            If aByte(Col) < 32 Or aByte(Col) > 126 Then
              str = str & "<" & aByte(Col) & ">"
            Else
              str = str & Chr(aByte(Col))
            End If
            Col = Col + 1
          Loop Until Col > UBound(aByte)
          Col = 1
          Do
            Sheets("Full>").Cells(Row, Col) = Mid(str, Col * 100 - 99, 100)
            Col = Col + 1
          Loop Until Col > Int(Len(str) / 100)
          Sheets("Full>").Cells(Row, Col) = Right(str, Len(str) - (Col - 1) * 100)
'          Col = 0
'          Do
'            i = aByte(Col) 'Fill in Character Count
'            Sheets("Main").Cells(i + 1, 1) = Sheets("Main").Cells(i + 1, 1) + 1
'            Col = Col + 1
'          Loop Until Col > UBound(aByte)
        End If
        Col = 1
        Sheets("Fields").Cells(Row, Col) = DateIn(aFile(iFile)) 'Date/Time (A)
        Col = Col + 1
        Sheets("Fields").Cells(Row, Col) = Round((DateIn(aText(7)) - Cells(Row, 1)) * 24 * 3600, 0) 'Duration in seconds (B)
        Do
          Col = Col + 1
          Sheets("Fields").Cells(Row, Col) = aText(Col - 3)
        Loop While Col < UBound(aStart) + 3
        i = Col + 1
        Sheets("Fields").Cells(1, Col + 1) = "FL fields"
        Do
          Col = Col + 1
          Sheets("Fields").Cells(Row, Col) = aFL(Col - i) 'Print Favorites List
        Loop Until Col - i > UBound(aFL) - 1
        i = Col + 1
        Sheets("Fields").Cells(1, Col + 1) = "System fields"
        Do
          Col = Col + 1
          Sheets("Fields").Cells(Row, Col) = aSys(Col - i) 'Print System
        Loop Until Col - i > UBound(aSys) - 1
        i = Col + 1
        Sheets("Fields").Cells(1, Col + 1) = "Department fields"
        Do
          Col = Col + 1
          Sheets("Fields").Cells(Row, Col) = aDept(Col - i) 'Print Departments
        Loop Until Col - i > UBound(aDept) - 1
        i = Col + 1
        Sheets("Fields").Cells(1, Col + 1) = "Channel fields"
        Do
          Col = Col + 1
          Sheets("Fields").Cells(Row, Col) = aChan(Col - i) 'Print Channels
        Loop Until Col - i > UBound(aChan) - 1
        i = Col + 1
        Sheets("Fields").Cells(1, Col + 1) = "Site fields"
        Do
          Col = Col + 1
          Sheets("Fields").Cells(Row, Col) = aSite(Col - i) 'Print Sites
        Loop Until Col - i > UBound(aSite) - 1
      End If
      '''''''''''''''''''''''''''''''''''''''' End Test Sheets ''''''''''''''''''''''''''''''''''''''''
      iFile = iFile + 1
    Loop
    iDir = iDir + 1
  Loop
  Cells.RowHeight = 12
  Cells.EntireColumn.AutoFit
  ''''''''''''''''''''''''''''''''''''''' Start Test Sheets '''''''''''''''''''''''''''''''''''''''
  If Test > 0 Then
    If Test > 1 Then
      Sheets("Full").Cells.EntireRow.AutoFit
      Sheets("Full").Cells.EntireColumn.AutoFit
      Sheets("Full>").Cells.EntireRow.AutoFit
      Sheets("Full>").Cells.EntireColumn.AutoFit
    End If
    Sheets("Fields").Cells.RowHeight = 12
    Sheets("Fields").Cells.EntireColumn.AutoFit
    Sheets("Fields").Cells.EntireColumn.AutoFit
  End If
  '''''''''''''''''''''''''''''''''''''''' End Test Sheets ''''''''''''''''''''''''''''''''''''''''
  Application.ScreenUpdating = True
End Sub
```

##### Importer

```vba
Option Explicit
'© 2014 Timothy H. Heaton, theaton@usd.edu, version 0.3
Private Sub FileList_DblClick(ByVal Cancel As MSForms.ReturnBoolean)
  OpenFile_Click
End Sub
Private Sub FileList_KeyPress(ByVal KeyAscii As MSForms.ReturnInteger)
  If KeyAscii = 27 Then
    Quit_Click
  ElseIf KeyAscii = 13 And Me!FileList <> "" Then
    OpenFile_Click
  End If
End Sub
Private Sub Ftype_Change()
  UserForm_Initialize
End Sub
Private Sub MyComp_Click()
  MyDir = "<.. My Computer>"
  UserForm_Initialize
End Sub
Private Sub MyDesk_Click()
  GrantAccessToMultipleFiles (Array("/Users/robbiet480/XLS_Audio/2020-06-21_12-16-36.wav"))
  If User = "" Then
'    MyDir = "C:\Documents and Settings"
    MyDir = "/Users/robbiet480/XLS_Audio/2020-06-21_12-16-36.wav"
  Else
    MyDir = User & "\Desktop"
  End If
  UserForm_Initialize
End Sub
Private Sub MyDoc_Click()
  If User = "" Then
'    MyDir = "C:\Documents and Settings"
    MyDir = "C:\Users"
  Else
'    MyDir = User & "\My Documents"
    MyDir = User & "\Documents"
  End If
  UserForm_Initialize
End Sub
Private Sub OpenFile_Click()
  If IsNull(Me!FileList) Then
    MsgBox "No file was selected.", , "  Error"
  ElseIf Me!FileList Like "?:" Then
    MyDir = Me!FileList
    UserForm_Initialize
  ElseIf Me!FileList Like "<*" Then
    If Me!FileList Like "<.. *" Then
      If InStrRev(MyDir, "\") = 0 Then
        MyDir = "<.. My Computer>"
      Else
        MyDir = Left(MyDir, InStrRev(MyDir, "\") - 1)
      End If
    ElseIf Me!FileList Like "<. *" Then
    Else
      MyDir = MyDir & "\" & Right(Left(Me!FileList, Len(Me!FileList) - 1), Len(Me!FileList) - 2)
    End If
    UserForm_Initialize
  Else
    Root = MyDir & "\" & Me!FileList
    ImportForm.Hide
  End If
End Sub
Private Sub Quit_Click()
  UserForm_Terminate
  ImportForm.Hide
End Sub
Private Sub UserForm_Initialize()
  Dim sd%, MyFile$
  Me!FileList.SetFocus
  If MyDir = "" Then MyDir = CurDir
  If User = "" Then
    If MyDir Like "?:\Users\?*\*" Then
      User = Left(MyDir, InStr(17, CurDir("C"), "\") - 1)
    ElseIf MyDir Like "?:\Userss\?*" Then
      User = MyDir
    ElseIf CurDir("C") Like "C:\Users\?*\*" Then
      User = Left(CurDir("C"), InStr(17, CurDir("C"), "\") - 1)
    ElseIf CurDir("C") Like "C:\Users\?*" Then
      User = CurDir("C")
    ElseIf MyDir Like "?:\Documents and Settings\?*\*" Then
      User = Left(MyDir, InStr(27, CurDir("C"), "\") - 1)
    ElseIf MyDir Like "?:\Documents and Settings\?*" Then
      User = MyDir
    ElseIf CurDir("C") Like "C:\Documents and Settings\?*\*" Then
      User = Left(CurDir("C"), InStr(27, CurDir("C"), "\") - 1)
    ElseIf CurDir("C") Like "C:\Documents and Settings\?*" Then
      User = CurDir("C")
    End If
  End If
  If Me!Ftype.ListCount = 0 Then
    Me!Ftype.AddItem "All files"
    Me!Ftype.AddItem "WAV files"
    Me!Ftype.ListIndex = 1
  End If
  On Error Resume Next
  Set fs = CreateObject("Scripting.FileSystemObject")
  If MyDir = "<.. My Computer>" Then
    Me!FullDir = "My Computer"
    Dim d, dc
    Set dc = fs.Drives
    Me!FileList.Clear
    For Each d In dc
      Me!FileList.AddItem d.DriveLetter & ":"
      If d.DriveType = 3 Then
        Me!FileList.List(Me!FileList.ListCount - 1, 2) = d.ShareName
      Else
        Me!FileList.List(Me!FileList.ListCount - 1, 2) = d.VolumeName
        If Err.Number > 0 Then 'Empty disk drive generates error
          Err.Clear
          Me!FileList.List(Me!FileList.ListCount - 1, 2) = "Disk Drive (empty)"
        End If
      End If
    Next
    Exit Sub
  End If
  MyFile = Dir(MyDir & "\*.*", vbDirectory) 'Reads first file in directory (Error on empty drive)
  If MyFile = "" Or Err.Number > 0 Then
    MsgBox MyDir & " could not be opened." & vbLf & "Try selecting another drive or directory." _
      & vbLf & "You may need to save and reopen the workbook.", , "  Error"
    Exit Sub
  End If
  Me!FileList.Clear
  Do Until Len(MyFile) = 0 Or i = 100
    If MyFile = "." Then
      If Len(MyDir) > 2 Then 'Skip drive letter on network drives
        Me!FileList.AddItem "<. " & Right(MyDir, Len(MyDir) - InStrRev(MyDir, "\")) & ">"
        Me!FileList.List(Me!FileList.ListCount - 1, 2) = "Current directory"
      End If
    ElseIf MyFile = ".." Then
      If Len(MyDir) > 2 Then 'Skip drive letter on network drives
        sd = InStrRev(MyDir, "\") - 1
        Me!FileList.AddItem "<.. " & Right(Left(MyDir, sd), sd - InStrRev(MyDir, "\", sd)) & ">", 0
        Me!FileList.List(0, 2) = "Parent directory"
        sd = 2
      End If
    ElseIf GetAttr(MyDir & "\" & MyFile) And 16 Then
      Me!FileList.AddItem "<" & MyFile & ">", sd
      Me!FileList.List(sd, 2) = "Subdirectory"
      sd = sd + 1
    ElseIf Me!Ftype = "All files" Or LCase(MyFile) Like "*." & LCase(Left(Me!Ftype, InStr(Me!Ftype, " ") - 1)) Then
      Set f = fs.GetFile(MyDir & "\" & MyFile)
      Me!FileList.AddItem MyFile
      Me!FileList.List(Me!FileList.ListCount - 1, 1) = CStr(f.Size)
      Me!FileList.List(Me!FileList.ListCount - 1, 2) = f.DateLastModified
    End If
    MyFile = Dir() 'Reads subsequent files in directory
  Loop
  If MyDir Like "?:" Then
    Me!FileList.AddItem "<. " & MyDir & ">", 0
    Me!FileList.List(0, 2) = "Current directory"
    Me!FileList.AddItem "<.. My Computer>", 0
    Me!FileList.List(0, 2) = "Parent directory"
    sd = sd + 2
  End If
  If Me!FileList.ListCount - sd = 0 Then
    Me!FullDir = MyDir & "\"
  Else
    If Me!FileList.ListCount - sd = 1 Then
      MyFile = " file)"
    Else
      MyFile = " files)"
    End If
    Me!FullDir = MyDir & "\  (" & Me!FileList.ListCount - sd & MyFile
    Me!FileList.List(0, 1) = "bytes"
  End If
End Sub
Private Sub UserForm_Terminate()
'  Root = ""
End Sub
```
