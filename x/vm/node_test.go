package xvm

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"gorgonia.org/gorgonia"
)

func Test_receiveInput(t *testing.T) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	inputC := make(chan ioValue, 0)
	type args struct {
		ctx context.Context
		o   *node
		fn  func()
	}
	tests := []struct {
		name string
		args args
		want stateFn
	}{
		{
			"context cancelation",
			args{
				cancelCtx,
				&node{},
				nil,
			},
			nil,
		},
		{
			"bad input value position",
			args{
				context.Background(),
				&node{
					inputC:      inputC,
					inputValues: make([]gorgonia.Value, 1),
				},
				func() {
					inputC <- struct {
						pos int
						v   gorgonia.Value
					}{
						pos: 1,
						v:   nil,
					}
				},
			},
			nil,
		},
		{
			"more value to receive",
			args{
				context.Background(),
				&node{
					inputC:      inputC,
					inputValues: make([]gorgonia.Value, 2),
				},
				func() {
					inputC <- struct {
						pos int
						v   gorgonia.Value
					}{
						pos: 0,
						v:   nil,
					}
				},
			},
			receiveInput,
		},
		{
			"all done go to compute",
			args{
				context.Background(),
				&node{
					inputC:      inputC,
					inputValues: make([]gorgonia.Value, 1),
				},
				func() {
					inputC <- struct {
						pos int
						v   gorgonia.Value
					}{
						pos: 0,
						v:   nil,
					}
				},
			},
			computeFwd,
		},
		// TODO: Add test cases.
	}
	cancel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.fn != nil {
				go tt.args.fn()
			}
			got := receiveInput(tt.args.ctx, tt.args.o)
			gotPrt := reflect.ValueOf(got).Pointer()
			wantPtr := reflect.ValueOf(tt.want).Pointer()
			if gotPrt != wantPtr {
				t.Errorf("receiveInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_computeFwd(t *testing.T) {
	type args struct {
		in0 context.Context
		n   *node
	}
	tests := []struct {
		name string
		args args
		want stateFn
	}{
		{
			"simple no error",
			args{
				nil,
				&node{
					op:          &noOpTest{},
					inputValues: []gorgonia.Value{nil},
				},
			},
			emitOutput,
		},
		{
			"simple with error",
			args{
				nil,
				&node{
					op:          &noOpTest{err: errors.New("")},
					inputValues: []gorgonia.Value{nil},
				},
			},
			nil,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeFwd(tt.args.in0, tt.args.n)
			gotPrt := reflect.ValueOf(got).Pointer()
			wantPtr := reflect.ValueOf(tt.want).Pointer()
			if gotPrt != wantPtr {
				t.Errorf("computeFwd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_node_ComputeForward(t *testing.T) {
	type fields struct {
		op             gorgonia.Op
		output         gorgonia.Value
		outputC        chan gorgonia.Value
		receivedValues int
		err            error
		inputValues    []gorgonia.Value
		inputC         chan ioValue
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"simple",
			fields{
				op: nil,
			},
			args{
				nil,
			},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{
				op:             tt.fields.op,
				output:         tt.fields.output,
				outputC:        tt.fields.outputC,
				receivedValues: tt.fields.receivedValues,
				err:            tt.fields.err,
				inputValues:    tt.fields.inputValues,
				inputC:         tt.fields.inputC,
			}
			if err := n.Compute(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("node.ComputeForward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type sumF32 struct{}

func (*sumF32) Do(v ...gorgonia.Value) (gorgonia.Value, error) {
	val := v[0].Data().(float32) + v[1].Data().(float32)
	value := gorgonia.F32(val)
	return &value, nil
}

func Test_emitOutput(t *testing.T) {
	cancelCtx, cancel := context.WithCancel(context.Background())
	outputC := make(chan gorgonia.Value, 1)
	type args struct {
		ctx context.Context
		n   *node
	}
	tests := []struct {
		name string
		args args
		want stateFn
	}{
		{
			"context cancelation",
			args{
				cancelCtx,
				&node{},
			},
			nil,
		},
		{
			"emit output",
			args{
				context.Background(),
				&node{
					outputC: outputC,
				},
			},
			nil,
		},
	}
	cancel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emitOutput(tt.args.ctx, tt.args.n)
			gotPrt := reflect.ValueOf(got).Pointer()
			wantPtr := reflect.ValueOf(tt.want).Pointer()
			if gotPrt != wantPtr {
				t.Errorf("emitOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_computeBackward(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 *node
	}
	tests := []struct {
		name string
		args args
		want stateFn
	}{
		{
			"simple",
			args{
				nil,
				nil,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeBackward(tt.args.in0, tt.args.in1); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeBackward() = %v, want %v", got, tt.want)
			}
		})
	}
}
