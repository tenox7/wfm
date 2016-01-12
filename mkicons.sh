
cat /dev/null > wfmiconres.h
echo "if (0) ;" > wfmicondis.h

for file in icons/*.gif
do
	./bin2c -c $file wfmiconres.h
	filename=$(basename $file)
	filedef=$(echo $filename | tr "." "_")
	echo "else if(strcmp(icon_name, \"${filename}\")==0) fwrite(${filedef}, sizeof(${filedef}), 1, cgiOut);" >> wfmicondis.h
done
