CC=gcc -Wall -O3 -DWFMGIT 
OB=wfm.o dir.o dialogs.o fileio.o cgic.o md5.o urlencode.o git.o
LD=-lgit2 -lpthread

all: wfm

wfm: ${OB}
	${CC} ${OB} -o wfm ${LD}
	@strip wfm
	@du -h wfm

wfm.o: wfm.c wfmiconres.h wfm.h

wfmiconres.h: bin2c
	sh ./mkicons.sh

bin2c: bin2c.c
	${CC} -o bin2c bin2c.c

.c.o:
	${CC} -c $< 

clean:
	rm -f *.o  wfm wfmicon*.h bin2c

