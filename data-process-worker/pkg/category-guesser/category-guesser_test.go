package categoryguesser

import "testing"

func TestGuessCategory(t *testing.T) {
	type args struct {
		transactionName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Should return Gym",
			args: args{
				transactionName: "FITNESS WORLD GEORGIA",
			},
			want:    "Gym",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: " SCOTIABANK TRANSIT 01420      ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return Credit Card",
			args: args{
				transactionName: "FROM - *****23*1986 ",
			},
			want:    "Credit Card",
			wantErr: false,
		},
		{
			name: "Should return Gym",
			args: args{
				transactionName: "FITNESS WORLD GEORGIA",
			},
			want:    "Gym",
			wantErr: false,
		},
		{
			name: "Should return Phone",
			args: args{
				transactionName: "VIRGIN PLUS    VERDUN       QC ",
			},
			want:    "Phone",
			wantErr: false,
		},
		{
			name: "Should return Gym",
			args: args{
				transactionName: "FITNESS WORLD GEORGIA",
			},
			want:    "Gym",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "TREES CHEESE CAKE AND ORG       ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "HM Pacific Centre        Vancouver       ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return Gym",
			args: args{
				transactionName: "FITNESS WORLD GEORGIA  ",
			},
			want:    "Gym",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "LONDON DRUGS 02       ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "MARKETPLACE IGA # 16  ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return Uber",
			args: args{
				transactionName: "UBER CANADA/UBERTRIP ON   ",
			},
			want:    "Uber",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "DISTRICT FACTORY OUTLET",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "Five Guys 1594 Vancouver       ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "#337 SPORT CHEK       ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "MARKETPLACE IGA # 16  ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "Riot* AN3SJF00TZJW       866-373-9211 CA ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "MARKETPLACE IGA # 16  ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "AMZN Mktp CA*TL6BJ7MG1   WWW.AMAZON.CAON ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "7 ELEVEN STORE #35667 ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return Phone",
			args: args{
				transactionName: "VIRGIN PLUS    VERDUN       QC ",
			},
			want:    "Phone",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "Amazon.ca*TR4FJ2410      AMAZON.CA    ON ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "Amazon.ca*TR5UC34O0      AMAZON.CA    ON ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "AMZN Mktp CA*TR8H413T0   WWW.AMAZON.CAON ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "GRETA BAR YVR",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "Amazon.ca*TL35X8U51      AMAZON.CA    ON ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return Uber",
			args: args{
				transactionName: "UBER* TRIP ON   ",
			},
			want:    "Uber",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "7 ELEVEN STORE #35667 ",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "MONZO BURGER",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Should return empty string",
			args: args{
				transactionName: "7 ELEVEN STORE #35667 ",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GuessCategory(tt.args.transactionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GuessCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GuessCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}
