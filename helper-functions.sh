function here {
  printf "\n\n" && pwd && printf "\n\n"
}

function countHistory {
  sortHistory
  sort -b -f ~/.bash_history | uniq -c
}

function sortHistory {
  history -w ~/Documents/bash-history_tmp.txt
  cat ~/Documents/bash-history_tmp.txt >> ~/Documents/bash-history.txt
  cat ~/Documents/bash-history.txt | awk '{$1=$1;print}' |  sort -b -f -u -o ~/Documents/sortedHistory_awk.txt
  printf "\nUnique Commands in History List:"
  cat ~/Documents/sortedHistory_awk.txt | wc -l
  cp ~/Documents/sortedHistory_awk.txt ~/Documents/bash-history.txt
  printf "\n"
}

function searchHistory {
  sortHistory
  grep -i $1 ~/Documents/sortedHistory_awk.txt
}

function showHistory {
  sortHistory
  cat ~/Documents/sortedHistory_awk.txt
}

HISTFILESIZE=10000