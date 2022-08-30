package injector

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHandler_ShouldInject(t *testing.T) {
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "should return false on no annotations",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "should return false if injection annotation is present but disabled",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/inject-agent": "false",
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "should return true if injection annotation is present and enabled",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/inject-agent": "true",
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "should return error if injection annotation is present but has invalid value",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/inject-agent": "asdf",
						},
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "should return true on if injection is enabled and agent status is not injected",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/inject-agent": "true",
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "should return false on explicit deny",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/inject-agent": "true",
							"heist.youniqx.com/agent-status": "injected",
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "should return false if injected is disabled but the agent is also already injected",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/inject-agent": "false",
							"heist.youniqx.com/agent-status": "injected",
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{
				Filter: operator.NewFilterWithValue(""),
				Log:    logr.Discard(),
			}
			got, err := handler.ShouldInject(tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShouldInject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ShouldInject() got = %v, want %v", got, tt.want)
			}
		})
	}
}
