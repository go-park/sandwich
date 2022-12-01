package aspectlib

type Annoation string

const (
	// CommentProxy for struct while comment @Proxy generate a file with _gen.go suffix
	CommentProxy Annoation = "@Proxy"
	// CommentPointcut for struct function while comment @Pointcut generate a proxy func for proxy struct
	CommentPointcut Annoation = "@Pointcut"
	// CommentAspect for struct while comment @Aspect then use to enhance other function
	CommentAspect Annoation = "@Aspect"
	// CommentAdviceBefore for struct function while comment @Before then use to enhance other function
	CommentAdviceBefore Annoation = "@Before"
	// CommentAdviceAfter for struct function while comment @After then use to enhance other function
	CommentAdviceAfter Annoation = "@After"
	// CommentAdviceAround for struct function while comment @Around then use to enhance other function
	CommentAdviceAround Annoation = "@Around"
)

var (
	adviceAnnoationList = []Annoation{CommentAdviceBefore, CommentAdviceAfter, CommentAdviceAround}
	funcAnnoationList   = append(adviceAnnoationList, CommentPointcut)
)

func (a Annoation) String() string {
	return string(a)
}
