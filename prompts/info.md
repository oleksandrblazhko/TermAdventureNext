To alleviate the workload spent on creating the same level with
just a few different words, TermAdventure also provides templat-
ing capabilities. Templates are created in the same way and format
as challenge definition files. The template contains static parts,
usually the majority of the text, as well as some special syntax
describing how dynamic content will be inserted. It utilizes Go’s
html/templates package as a template engine. This package
is normally used for HTML templating, is therefore well tested and
contains many features relevant for our use case (such as loops,
variables or filters). To create a template, build a normal challenge
definition file, with the dynamic part written in standard Go tem-
plate language, such as using {{, }} to mark dynamic variables.

Values for the template variables go to a separate yaml file con-
taining key-value pairs, where the key is a name of the variable in-
side the template file. A simple template with its accompanying vari-
able yaml file can be seen in Listing 2 and 3, respectively. TermAd-
venture comes with few predefined filters, like generate_levels,
which is used to generate a list of upcoming levels for the next

keyword as seen in Listing 2. It takes an iterable from the vari-
ables file as an argument together with the name format of the

levels. For example, in Listing 2 the output of the filter would be
['lvl21', 'lvl22', 'lvl23'] .
All of these features create a simple and a straightforward way
of creating a adventure, all of which can be done just from the
command line itself. For this reason a TermAdventure challenge
can run even on GUI-less environment, such as servers, and is very
portable from one environment to the other. Furthermore, since the
adventures use only bash capabilities, such as the test command,
once the adventure is created it can run on most of the modern
UNIX architectures.
Listing 2: An example of a ”adventure” template file.
1 name: lvl1
2 next: [{{generate_levels "lvl2" $.dirs "%s%d"}}]
3 test: /usr/bin/test $(pwd) = "/tmp"
4
5 In this level your task is to change your
6 working directory to *'/tmp'*.
7
8 --------------------
9 {{ range $index, $dir := .directories }}
10 name: lvl2{{ add $index 1 }}
11 next: ['finish']
12 test: /usr/bin/test $(pwd) = "{{ dir }}"
13
14 Awesome, you made it to **/tmp**. Now get
15 to the */{{ dir }}* directory.
16
17 --------------------
18 {{ end }}
19 name: finish

ITiCSE 2021, June 26-July 1, 2021, Virtual Event, Germany Šuppa and Jariabka, et al.
20 test: false
21
22 # I see you made it again, awesome!
23 That's all for now, so lay back and enjoy
24 your shell!
25
26 --------------------