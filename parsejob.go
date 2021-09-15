/*
	This is a golang rewrite of my Python 3 program that reads the
	Hercules prt00e, prt00f and prt002 text files from MVS
	(tk4- in my case) which contains multiple jobs, back to back and
	parses the jobs out to individual text files. It is meant to run
	within the Hercules "prt" subdirectory.

	By: Bill Blasingim
	On: 09/13/2021

	20210914 - Simplified input file search. Remove searching directory
		for Mainframe queue files. Replace with simple array checking
		for CUUs 002, 00e, 00f. Create the subdirectories if they don't exist!
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func process(ifil string, odir string) {
	fmt.Println("Processing", ifil, odir)
	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	cwd = cwd + "/" + odir + "/"
	outdir := cwd

	alpha := "A"
	if ifil[5:6] == "f" {
		alpha = "Z"
	}

	infile, err := os.Open(ifil)
	if err != nil {
		fmt.Println(err)
	}
	defer infile.Close()

	scanner := bufio.NewScanner(infile)
	/*
		#########################################
		#                Read Loop              #
		#########################################
	*/
	cnt := 0
	ofile := ""
	endCnt := 0
	tfil := ""
	OpenTmp := true
	/*
		The below outfil var is how I, as an amateur in Go handled a
		problem I was having refrencing outfil later in the program. There
		may be a better way that is obvious to a more experience go programmer.
		Normally you would do this "outfil, err := os.Create(tfil)"
		(note the":=") but evidenly I needed to make it global, because the
		program wasn't aware of outfil down further. So I decided to make this
		var global.I got the outfil type by using the %T in fmt.Printf
	*/

	var outfil *os.File
	for scanner.Scan() {
		if OpenTmp {
			//fmt.Println("Opening Temp!")

			/*
				Start out writing temp file since I won't know output file name
				till I get to bottom of 1st page
			*/
			dt := time.Now()

			// printing the time in string format
			tfil = dt.String() + ".tmp"
			tfil = cwd + strings.Replace(tfil, " ", "", -1)
			outfil, err = os.Create(tfil)
			if err != nil {
				log.Fatal(err)
			}
			OpenTmp = false
		}

		txt := scanner.Text()
		LineOut := txt

		cnt++

		ln := len(txt)
		job := ""
		jobName := ""
		timeDate := ""

		if ln > 20 && LineOut[0:12] == "****"+alpha+"  START" {
			endCnt = 0
			job = LineOut[17:23]
			jobName = LineOut[24:33]
			timeDate = LineOut[67:88]
			// ofil is the filename I will rename the current temp file to
			ofile = outdir + strings.Trim(job, " ") + "-" + strings.Trim(jobName, " ") + " (" + strings.Trim(timeDate, " ") + ").txt"
		}

		outfil.WriteString(LineOut + "\n")

		if ln > 20 && LineOut[0:11] == "****"+alpha+"   END" {
			endCnt++
			if endCnt > 3 {
				outfil.Close()
				//fmt.Println("From : ", tfil) //, " >> ", ofile) //<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
				//fmt.Println("To   : ", ofile, "\n")
				e := os.Rename(tfil, ofile)
				if e != nil {
					log.Fatal(e)
				}
				fmt.Println("Created", ofile)
				OpenTmp = true

			}

		}

	} // End of for loop

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	// Delete last temp fil created in anticipation of a new Job
	if tfil != "" {
		e := os.Remove(tfil)
		if e != nil {
			log.Fatal(e)
		}
	}
}

/*
	#########################################
	#                M A I N                #
	#########################################
*/
func main() {
	cuu := []string{"002", "00e", "00f"}
	for i := 0; i < 3; i++ {
		odir := "prt" + cuu[i]
		_, err := os.Stat(odir)
		if err != nil {
			fmt.Println("Creating", odir)
			err2 := os.MkdirAll(odir, 0755)

			if err2 != nil {
				log.Println("Error creating directory")
				log.Println(err)
				return
			}
		}
		ifil := odir + ".txt"
		fmt.Println("Looking for:", ifil)

		process(ifil, odir)
	}

	fmt.Println("Completed!")

}
