#!/bin/bash
touch out
for i in *.wav; do

# IART-System name begins in position 33 and is 64 characters wide
dd skip=32 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# IGND - Department name
dd skip=104 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# INAM - Channel name
dd skip=176 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# ICMT - Mystery number
dd skip=248 count=8 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# IPRD - Scanner name
dd skip=320 count=16 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# IKEY - realted to system name
dd skip=346 count=16 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# ICRD - Closing date/time
dd skip=376 count=12 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# ISRC - Tone or NAC
dd skip=400 count=16 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# ITCH - Unit ID
dd skip=424 count=16 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# ISBJ - Favorite List Name
dd skip=497 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# ICOP
dd skip=568 count=16 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# Favorite List block
dd skip=592 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# System Block
dd skip=657 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# Department block
dd skip=722 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# Channel Block
dd skip=787 count=64 bs=1 if=$i | tr -cd '\40-\176' >>out
echo -n "," >>out
# Site block
dd skip=852 count=64 bs=1 if=$i| tr -cd '\40-\176' >>out
echo -n "," >>out
# TGID
dd skip=917 count=16 bs=1 if=$i| tr -cd '\40-\176' >>out
echo " " >>out

done
